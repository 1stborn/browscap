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

type bsMode int

const (
	Lite bsMode = iota
	Full
)

const bsVersion = "http://Browscap.org/version"
const bsStream = "https://Browscap.org/stream?q="

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

	mode bsMode

	browsers, platforms map[uint32]string

	defaults map[uint32]Browser
	tree     *browserTree

	m sync.RWMutex
}

func (bm bsMode) ServiceCached(dir string, fn func(Version)) *Browscap {
	return bm.startService(func(release time.Time) (io.ReadCloser, error) {
		return new(readProxy).Proxy(
			dir+"/bs_"+release.Format("20060102150405")+".csv", func() (io.ReadCloser, error) {
				return bm.readUpstream(release)
			})
	}, fn)
}

func (bm bsMode) Service(fn func(Version)) *Browscap {
	return bm.startService(bm.readUpstream, fn)
}

func (bm bsMode) startService(reader func(time.Time) (io.ReadCloser, error), fn func(Version)) *Browscap {
	var last time.Time

	b := bm.new()

	update := func(now time.Time) {
		if now.Sub(last) > time.Hour {
			if resp, err := http.Get(bsVersion); err == nil && resp.StatusCode == 200 {
				last = now

				bytes, _ := ioutil.ReadAll(resp.Body)
				release, _ := time.Parse(time.RFC1123Z, string(bytes))

				if b.Time.Before(release) {
					if r, err := reader(release); err == nil {
						b.fromCsv(csv.NewReader(r))
						if fn != nil {
							fn(b.Version)
						}

						r.Close()
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

func (bm bsMode) readUpstream(time.Time) (io.ReadCloser, error) {
	if resp, err := http.Get(bsStream + defaultStream); err == nil {
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
	var weight int
	for {
		if record, err := reader.Read(); err == nil {
			weight++
			bs.Version.Count += bs.add(record, weight)
		} else if err == io.EOF {
			break
		}
	}
}

func (bs *Browscap) Count() int {
	return bs.Version.Count
}

func (bm bsMode) new() *Browscap {
	b := &Browscap{mode: bm, tree: newTree()}

	b.defaults = make(map[uint32]Browser)
	b.platforms = make(map[uint32]string)
	b.browsers = make(map[uint32]string)

	return b
}

func (bm bsMode) Csv(file string) *Browscap {
	if f, err := os.OpenFile(file, os.O_RDONLY, 0); err == nil {
		return bm.CsvReader(csv.NewReader(f))
	}

	return nil
}

func (bm bsMode) CsvReader(reader *csv.Reader) *Browscap {
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

func (bs *Browscap) add(opts []string, weight int) int {
	if len(opts) > 50 {
		if fMasterParent.Is(opts) {
			br := Browser{bs: bs}
			br.mapArray(opts)

			_, hash := fPropertyName.Hash(opts)

			bs.defaults[hash] = br

			return 0
		} else if bs.mode == Lite && !fLiteMode.Is(opts) {
			return 0
		} else {
			_, hash := fParent.Hash(opts)

			if br, ok := bs.defaults[hash]; ok {
				bs.tree.Add(opts, &br)
			} else {
				bs.tree.Add(opts, &Browser{bs: bs})
			}
		}

		if fMasterParent.Is(opts) {
			return 0
		}

		if fIsFake.Is(opts) {
			return 0
		}

		if bs.mode == Lite && !fLiteMode.Is(opts) {
			return 0
		}

		bs.tree.Add(opts, &Browser{bs: bs})

		return 1
	}

	return 0
}

func (proxy *readProxy) Proxy(file string, fn func() (io.ReadCloser, error)) (io.ReadCloser, error) {
	if f, err := os.OpenFile(file, os.O_RDONLY, 0); err == nil {
		return f, nil
	} else if r, err := fn(); err == nil {
		if f, err := os.OpenFile(file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0); err == nil {
			pr, pw := io.Pipe()

			go func() {
				var buffer = make([]byte, 4096)

				for {
					if n, err := r.Read(buffer); err != io.EOF {
						f.Write(buffer[:n])
						pw.Write(buffer[:n])
					} else {
						pw.Close()

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