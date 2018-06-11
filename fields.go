package browscap

import (
	"strconv"
	"reflect"
	"unsafe"
	"hash/crc32"
)

var crc = crc32.NewIEEE()

type browscapField int

const (
	fPropertyName                browscapField = iota
	fMasterParent
	fLiteMode
	fParent
	fComment
	fBrowser
	fBrowserType
	fBrowserBits
	fBrowser_Maker
	fBrowser_Modus
	fVersion
	fMajorVer
	fMinorVer
	fPlatform
	fPlatform_Version
	fPlatform_Description
	fPlatformBits
	fPlatform_Maker
	fAlpha
	fBeta
	fWin16
	fWin32
	fWin64
	fFrames
	fIFrames
	fTables
	fCookies
	fBackgroundSounds
	fJavaScript
	fVBScript
	fJavaApplets
	fActiveXControls
	fIsMobileDevice
	fIsTablet
	fisSyndicationReader
	fCrawler
	fIsFake
	fisAnonymized
	fisModified
	fCssVersion
	fAolVersion
	fDevice_Name
	fDevice_Maker
	fDeviceType
	fDevicePointingMethod
	fDevice_Code_Name
	fDevice_Brand_Name
	fRenderingEngine_Name
	fRenderingEngine_Version
	fRenderingEngine_Description
	fRenderingEngine_Maker
)

func (nf browscapField) Is(opts []string) bool {
	if len(opts) > int(nf) {
		if ok, _ := strconv.ParseBool(opts[int(nf)]); ok {
			return true
		}
	}

	return false
}

func (nf browscapField) Hash(opts []string) (string, uint32) {
	crc.Reset()

	s := nf.GetString(opts)
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{Data: sh.Data, Len: sh.Len, Cap: sh.Len}

	crc.Write(*(*[]byte)(unsafe.Pointer(&bh)))

	return s, crc.Sum32()
}

func (nf browscapField) Equals(opts []string, v string) bool {
	if len(opts) > int(nf) {
		return opts[int(nf)] == v
	}

	return false
}

func (nf browscapField) GetString(opts []string) string {
	if len(opts) > int(nf) {
		return opts[int(nf)]
	}

	return ""
}

func (nf browscapField) GetInt(opts []string) int {
	if v := nf.GetString(opts); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return 0
}
