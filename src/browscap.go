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
)

type browscapMode int

const (
	Lite browscapMode = iota
	Full
)

const browscapVersion = "http://browscap.org/version"
const browscapStream = "https://browscap.org/stream?q="

const defaultStream = "BrowsCapCSV"

type Version struct {
	Release int
	Time    time.Time
}

type browscap struct {
	Version

	mode browscapMode

	count            int
	browsers,platforms map[uint32]string

	defaults map[string]Browser
	tree     radixTree

	m sync.RWMutex
}


func (bm browscapMode) Service(fn func(Version)) *browscap {
	var last time.Time

	b := bm.new()

	update := func(now time.Time) {
		if now.Sub(last) > time.Hour {
			if resp, err := http.Get(browscapVersion); err == nil && resp.StatusCode == 200 {
				last = now

				bytes, _ := ioutil.ReadAll(resp.Body)
				release, _ := time.Parse(time.RFC1123Z, string(bytes))

				if b.Time.Before(release) {
					if resp, err := http.Get(browscapStream + defaultStream); err == nil {
						b.fromCsv(csv.NewReader(resp.Body))
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

func (b browscap) Find(agent string) *Browser {
	b.m.RLock()
	defer b.m.RUnlock()

	return b.tree.Find(agent)
}

func (b *browscap) readVersion(reader *csv.Reader) {
	reader.FieldsPerRecord = 0
	reader.Read() // consume header

	for {
		if v, err := reader.Read(); err == nil {
			b.Release, _ = strconv.Atoi(v[0])
			b.Time, _ = time.Parse(time.RFC1123Z, v[1])
		} else if _, ok := err.(*csv.ParseError); ok {
			break
		}
	}
}

func (b *browscap) readBrowsers(reader *csv.Reader) {
	reader.FieldsPerRecord = 0

	for {
		if record, err := reader.Read(); err == nil {
			b.count += b.add(record)
		} else if err == io.EOF {
			break
		}
	}
}

func (b *browscap) Count() int {
	return b.count
}

func (bm browscapMode) new() *browscap {
	b := &browscap{mode: bm}

	b.defaults = make(map[string]Browser)
	b.platforms = make(map[uint32]string)
	b.browsers = make(map[uint32]string)

	return b
}

func (bm browscapMode) Csv(file string) *browscap {
	if f, err := os.OpenFile(file, os.O_RDONLY, 0); err == nil {
		return bm.CsvReader(csv.NewReader(f))
	}

	return nil
}

func (bm browscapMode) CsvReader(reader *csv.Reader) *browscap {
	b := bm.new()
	b.fromCsv(reader)

	return b
}

func (b *browscap) fromCsv(reader *csv.Reader) {
	b.m.Lock()
	defer b.m.Unlock()

	b.readVersion(reader)
	b.readBrowsers(reader)
}

func (b *browscap) add(opts []string) int {
	if len(opts) > 50 {
		if fMasterParent.Is(opts) {
			br := Browser{browscap: b}
			br.mapArray(opts)
			b.defaults[fPropertyName.GetString(opts)] = br

			return 0
		} else if b.mode == Lite && !fLiteMode.Is(opts) {
			return 0
		} else if br, ok := b.defaults[fParent.GetString(opts)]; ok {
			b.tree.Add(opts, &br)
		} else {
			b.tree.Add(opts, &Browser{browscap: b})
		}

		return 1
	}

	return 0
}
