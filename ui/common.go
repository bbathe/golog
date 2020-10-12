package ui

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"unicode/utf8"
	"unsafe"

	"github.com/lxn/walk"
	"github.com/lxn/win"
)

// drawCellStyles paints the tableview cells how we want
func drawCellStyles(tv *walk.TableView, style *walk.CellStyle, sb walk.SorterBase) {
	// draw header cells ourselves
	if style.Row() == -1 {
		canvas := style.Canvas()
		cols := tv.Columns()
		col := style.Col()
		dpi := canvas.DPI()
		bounds := style.Bounds()

		// brush for cell background
		brush, err := walk.NewSolidColorBrush(walk.RGB(200, 200, 200))
		if err != nil {
			MsgError(nil, err)
			log.Printf("%+v", err)
			return
		}
		defer brush.Dispose()

		// pull back from the left
		b := walk.RectangleFrom96DPI(walk.Rectangle{
			X:      bounds.X,
			Y:      bounds.Y,
			Width:  bounds.Width - 1,
			Height: bounds.Height,
		}, dpi)

		err = canvas.FillRectanglePixels(brush, b)
		if err != nil {
			MsgError(nil, err)
			log.Printf("%+v", err)
			return
		}

		// font for header text
		f := tv.Font()
		font, err := walk.NewFont(f.Family(), f.PointSize(), walk.FontBold)
		if err != nil {
			MsgError(nil, err)
			log.Printf("%+v", err)
			return
		}

		b = walk.RectangleFrom96DPI(walk.Rectangle{
			X:      bounds.X + 4,
			Y:      bounds.Y + 1,
			Width:  bounds.Width - 8,
			Height: bounds.Height - 1,
		}, dpi)

		err = canvas.DrawTextPixels(cols.At(col).Title(), font, 0, b, walk.TextLeft)
		if err != nil {
			MsgError(nil, err)
			log.Printf("%+v", err)
			return
		}

		// draw sort indicator
		if sb.SortedColumn() == col {
			c := "\u2BC5"
			if sb.SortOrder() == walk.SortDescending {
				c = "\u2BC6"
			}
			err = canvas.DrawTextPixels(c, font, 0, b, walk.TextRight)
			if err != nil {
				MsgError(nil, err)
				log.Printf("%+v", err)
				return
			}
		}
	}
}

// formatFrequency returns a string with frequency formatted like on my IC-7300
func formatFrequency(freq string) string {
	// split on current decimal point
	parts := strings.Split(freq, ".")
	s1 := parts[0]

	// pad last part out to 2 digits
	s2 := parts[1] + strings.Repeat("0", 2-len(parts[1]))

	// figure out if we need to do anything (first part more than 3 digits)
	startOffset := 0
	const groupLen = 3
	groups := (len(s1) - startOffset - 1) / groupLen

	if groups == 0 {
		// recombine with formatted second part
		return s1 + "." + s2
	}

	sep := '.'
	sepLen := utf8.RuneLen(sep)
	sepBytes := make([]byte, sepLen)
	_ = utf8.EncodeRune(sepBytes, sep)

	buf := make([]byte, groups*(groupLen+sepLen)+len(s1)-(groups*groupLen))

	// move over in groups of 3, adding seperator
	startOffset += groupLen
	p := len(s1)
	q := len(buf)
	for p > startOffset {
		p -= groupLen
		q -= groupLen
		copy(buf[q:q+groupLen], s1[p:])
		q -= sepLen
		copy(buf[q:], sepBytes)
	}
	if q > 0 {
		copy(buf[:q], s1)
	}

	// recombine with formatted second part
	return string(buf) + "." + s2
}

// launchQRZPage opens the users default web browser to the qso partners QRZ.com page
func launchQRZPage(call string) error {
	u := "https://www.qrz.com"
	if call != "" {
		u += "/db/" + strings.Replace(call, "%", "", -1)
	}

	err := exec.Command(runDll32, "url.dll,FileProtocolHandler", u).Start()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

// launchPSKreporter opens the users default web browser to the qso partners PSKreporter report
func launchPSKreporter(call string) error {
	if call == "" {
		err := fmt.Errorf("must supply callsign")
		log.Printf("%+v", err)
		return err
	}
	u := fmt.Sprintf("https://www.pskreporter.de/?s_type=rcvdby&call=%s&search=Search", call)

	err := exec.Command(runDll32, "url.dll,FileProtocolHandler", u).Start()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

// copyToClipboard populates the clipboard with text txt
func copyToClipboard(txt string) error {
	// clear what's in there (if not then what we copy only occupies the plain text clipboard)
	err := walk.Clipboard().Clear()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	err = walk.Clipboard().SetText(txt)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

func flashWindow(hWnd win.HWND, times uint32) {
	type flashwinfo struct {
		CbSize    uint32
		Hwnd      win.HWND
		DwFlags   uint32
		UCount    uint32
		DwTimeout uint32
	}

	// flash both the window caption and taskbar button
	fw := flashwinfo{
		Hwnd:    hWnd,
		DwFlags: 0x00000003,
		UCount:  times,
	}
	fw.CbSize = uint32(unsafe.Sizeof(fw))

	// flash continuously until the window comes to the foreground?
	if times == 0 {
		fw.DwFlags |= 0x0000000C
	}

	_, _, _ = flashWindowEx.Call(uintptr(unsafe.Pointer(&fw)))
}
