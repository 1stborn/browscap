package browscap

import (
	"os"
	"encoding/csv"
	"time"
	"strconv"
	"io"
	"net/http"
	"io/ioutil"
	"sync"
	"fmt"
)

type browscapMode int

const (
	Lite browscapMode = iota
	Full
)

const browscapVersion = "http://Browscap.org/version"
const browscapStream = "https://Browscap.org/stream?q="

const defaultStream = "BrowsCapCSV"

type Version struct {
	Release int
	Time    time.Time
	Count   int
}

type readProxy struct {
	path   string
	stream func() (io.Reader, error)
}

type Browscap struct {
	Version

	mode browscapMode

	browsers, platforms map[uint32]string

	defaults map[string]Browser
	tree     radixTree

	m sync.RWMutex
}

func (bm browscapMode) ServiceCached(dir string, fn func(Version)) *Browscap {
	return bm.startService(func(release time.Time) (io.Reader, error) {
		return new(readProxy).Proxy(
			dir+"/bs_"+release.Format("20060102150405")+".csv", func() (io.Reader, error) {
				return bm.readUpstream(release)
			})
	}, fn)
}

func (bm browscapMode) Service(fn func(Version)) *Browscap {
	return bm.startService(bm.readUpstream, fn)
}

func (bm browscapMode) startService(reader func(time.Time) (io.Reader, error), fn func(Version)) *Browscap {
	var last time.Time

	b := bm.new()

	update := func(now time.Time) {
		if now.Sub(last) > time.Hour {
			if resp, err := http.Get(browscapVersion); err == nil && resp.StatusCode == 200 {
				last = now

				bytes, _ := ioutil.ReadAll(resp.Body)
				release, _ := time.Parse(time.RFC1123Z, string(bytes))

				if b.Time.Before(release) {
					if r, err := reader(release); err == nil {
						b.fromCsv(csv.NewReader(r))
						if fn != nil {
							fn(b.Version)
						}
					}
				}
			}
		}
	}

	update(time.Now())

	go func() {
		t := time.NewTicker(time.Minute)

		for {
			update(<-t.C)
		}
	}()

	return b
}

func (bm browscapMode) readUpstream(time.Time) (io.Reader, error) {
	if resp, err := http.Get(browscapStream + defaultStream); err == nil {
		return resp.Body, nil
	} else {
		return nil, err
	}
}

func (bs Browscap) Find(agent string) *Browser {
	bs.m.RLock()
	defer bs.m.RUnlock()

	return bs.tree.Find(agent)
}

func (bs *Browscap) readVersion(reader *csv.Reader) {
	reader.FieldsPerRecord = 0
	reader.Read() // consume header

	for {
		if v, err := reader.Read(); err == nil {
			bs.Release, _ = strconv.Atoi(v[0])
			bs.Time, _ = time.Parse(time.RFC1123Z, v[1])
		} else if _, ok := err.(*csv.ParseError); ok {
			break
		}
	}
}

func (bs *Browscap) readBrowsers(reader *csv.Reader) {
	reader.FieldsPerRecord = 0

	for {
		if record, err := reader.Read(); err == nil {
			bs.Version.Count += bs.add(record)
		} else if err == io.EOF {
			break
		}
	}
}

func (bs *Browscap) Count() int {
	return bs.Version.Count
}

func (bm browscapMode) new() *Browscap {
	b := &Browscap{mode: bm}

	b.defaults = make(map[string]Browser)
	b.platforms = make(map[uint32]string)
	b.browsers = make(map[uint32]string)

	return b
}

func (bm browscapMode) Csv(file string) *Browscap {
	if f, err := os.OpenFile(file, os.O_RDONLY, 0); err == nil {
		return bm.CsvReader(csv.NewReader(f))
	}

	return nil
}

func (bm browscapMode) CsvReader(reader *csv.Reader) *Browscap {
	b := bm.new()
	b.fromCsv(reader)

	return b
}

func (bs *Browscap) fromCsv(reader *csv.Reader) {
	bs.m.Lock()
	defer bs.m.Unlock()

	bs.readVersion(reader)
	bs.readBrowsers(reader)
}

func (bs *Browscap) add(opts []string) int {
	if len(opts) > 50 {
		if fMasterParent.Is(opts) {
			br := Browser{bs: bs}
			br.mapArray(opts)

			bs.defaults[fPropertyName.GetString(opts)] = br

			return 0
		} else if bs.mode == Lite && !fLiteMode.Is(opts) {
			return 0
		} else if br, ok := bs.defaults[fParent.GetString(opts)]; ok {
			bs.tree.Add(opts, &br)
		} else {
			bs.tree.Add(opts, &Browser{bs: bs})
		}

		return 1
	}

	return 0
}

func (proxy *readProxy) Proxy(file string, fn func() (io.Reader, error)) (io.Reader, error) {
	if f, err := os.OpenFile(file, os.O_RDONLY, 0); err == nil {
		return f, nil
	} else if r, err := fn(); err == nil {
		if f, err := os.OpenFile(file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0); err == nil {
			pr, pw := io.Pipe()

			go func() {
				var buffer = make([]byte, 4096)

				defer pr.Close()
				defer pw.Close()

				for {
					if n, err := r.Read(buffer); err != io.EOF {
						f.Write(buffer[:n])
						pw.Write(buffer[:n])
					} else {
						return
					}
				}
			}()

			return pr, nil
		} else {
			return r, nil
		}
	} else {
		return nil, err
	}
}

func (v Version) String() string {
	return fmt.Sprintf("%d@%s", v.Release, v.Time.Format("2006-01-02 15:04:05"))
}