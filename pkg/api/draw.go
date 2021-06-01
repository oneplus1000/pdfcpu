package api

import (
	"fmt"
	"os"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pkg/errors"
)

type Line struct {
	X1, Y1 float64
	X2, Y2 float64
}

func DrawLineFile(src string, dest string, lines []Line, conf *pdfcpu.Configuration) error {

	page := 1 //fake

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

	ctx, _, _, _, err := readValidateAndOptimize(f, conf, fromStart)
	if err != nil {
		return errors.Wrap(err, "readValidateAndOptimize fail")
	}
	_ = ctx

	consolidateRes := false
	dict, _, err := ctx.PageDict(page, consolidateRes)
	if err != nil {
		return err
	}
	obj, found := dict.Find("Contents")
	if !found {
		//จริงๆ ตรงนี้ต้องสรา้งใหม่ด้วย
		return nil
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
		err := drawLinesToStream(&sm)
		if err != nil {
			return err
		}
	} else if _, ok := obj.(pdfcpu.Array); ok {
		fmt.Print("Array")
	}

	return nil
}

func drawLinesToStream(sd *pdfcpu.StreamDict) error {
	err := sd.Decode()
	if err != nil {
		return err
	}
	return nil
}
