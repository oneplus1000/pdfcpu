package api

import (
	"fmt"
	"os"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pkg/errors"
)

//ErrDrawListNotSupport can not draw in pdf
var ErrDrawListNotSupport = errors.New("draw line not support")

type DrawLine struct {
	PageNumber int
	Lines      []Line
}
type Line struct {
	LineWidth float64
	X1, Y1    float64
	X2, Y2    float64
}

func DrawLineFile(src string, dest string, draws []DrawLine, conf *pdfcpu.Configuration) error {

	if conf == nil {
		conf = pdfcpu.NewDefaultConfiguration()
	}
	conf.Cmd = pdfcpu.DRAWLINES

	f, err := os.Open(src)
	if err != nil {
		return errors.Wrapf(err, "os.Open(%s) fail", src)
	}
	defer f.Close()

	var fromStart time.Time
	if conf.ForceCreationDate == nil {
		fromStart = time.Now()
	} else {
		fromStart = *conf.ForceCreationDate
	}

	ctx, _, _, err := readAndValidate(f, conf, fromStart)
	if err != nil {
		return errors.Wrap(err, "readValidateAndOptimize fail")
	}

	for _, d := range draws {
		err := drawLine(ctx, d)
		if err != nil {
			return errors.Wrap(err, "drawLine fail")
		}
	}

	fout, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return errors.Wrapf(err, "os.Open(%s) fail", dest)
	}
	defer fout.Close()

	if err = WriteContext(ctx, fout); err != nil {
		return err
	}

	return nil
}

func drawLine(ctx *pdfcpu.Context, dl DrawLine) error {
	consolidateRes := false
	dict, _, err := ctx.PageDict(dl.PageNumber, consolidateRes)
	if err != nil {
		return err
	}
	obj, found := dict.Find("Contents")
	if !found {
		return ErrDrawListNotSupport
	}

	var objNr int
	var entry *pdfcpu.XRefTableEntry
	if ir, ok := obj.(pdfcpu.IndirectRef); ok {
		objNr = ir.ObjectNumber.Value()
		genNr := ir.GenerationNumber.Value()
		entry, _ = ctx.FindTableEntry(objNr, genNr)
		obj = entry.Object
	}

	if sm, ok := obj.(pdfcpu.StreamDict); ok {
		err := drawLinesToStream(&sm, dl.Lines)
		if err != nil {
			return err
		}
		entry.Object = sm
	} else {
		return ErrDrawListNotSupport
	}

	return nil
}

func drawLinesToStream(sd *pdfcpu.StreamDict, lines []Line) error {
	err := sd.Decode()
	if err != nil {
		return err
	}
	for _, line := range lines {
		l0 := fmt.Sprintf("%.2f w\n", line.LineWidth)
		l1 := fmt.Sprintf("%0.2f %0.2f m %0.2f %0.2f l S\n", line.X1, line.Y1, line.X2, line.Y2)
		sd.Content = append(sd.Content, l0...)
		sd.Content = append(sd.Content, l1...)
	}
	return sd.Encode()
}
