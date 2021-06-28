package api

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pkg/errors"
)

const colorTypeStroke = "RG"

//const colorTypeFill = "rg"

//ErrDrawListNotSupport can not draw in pdf
var ErrDrawListNotSupport = errors.New("draw line not support")

type DrawLine struct {
	PageNumber int
	Lines      []Line
}
type Line struct {
	Alpha     float64 //0 to 1
	Color     ColorRGB
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

	conf.WriteXRefStream = false   //TODO: ลบด้วย
	conf.WriteObjectStream = false //TODO: ลบด้วย
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

	//debug, _ := ctx.FindTableEntry(13, 0)
	//debug.Compressed = false
	//fmt.Printf("DE %s\n", debug.Object.String())

	if err = WriteContext(ctx, fout); err != nil {
		return err
	}

	return nil
}

func setupExtGStateToRes(ctx *pdfcpu.Context, pageDict pdfcpu.Dict, mapOfAlphaAndIDs map[float64]string) error {
	//var entryRes *pdfcpu.XRefTableEntry
	var entryRes *pdfcpu.XRefTableEntry
	resDict, found := pageDict.Find("Resources")
	if found {
		if ir, ok := resDict.(pdfcpu.IndirectRef); ok {
			objNr := ir.ObjectNumber.Value()
			genNr := ir.GenerationNumber.Value()
			entryRes, _ = ctx.FindTableEntry(objNr, genNr)
			resDict = entryRes.Object
		}
	} else {
		//TODO: ถ้าไม่เจอเอาไง???
	}

	for key, _ := range mapOfAlphaAndIDs {
		extGState, _ := ctx.CreateExtGStateForTransparent(key)
		if rd, ok := resDict.(pdfcpu.Dict); ok {
			gsID := "GS0"
			o, found := rd.Find("ExtGState")
			if !found {
				rd.Insert("ExtGState", pdfcpu.Dict(map[string]pdfcpu.Object{gsID: *extGState}))
			} else {
				d, _ := ctx.DereferenceDict(o)
				for i := 0; i < 1000; i++ {
					gsID = "GS" + strconv.Itoa(i)
					_, found := d.Find(gsID)
					if !found {
						break
					}
				}
				d.Insert(gsID, *extGState)
			}
			if entryRes != nil {
				entryRes.Object = rd
			}
			mapOfAlphaAndIDs[key] = gsID
		}

	}
	//fmt.Printf("resDict : %+v\n", resDict)
	pageDict.Update("Resources", resDict)

	return nil
}

func drawLine(ctx *pdfcpu.Context, dl DrawLine) error {
	consolidateRes := false
	dict, _, err := ctx.PageDict(dl.PageNumber, consolidateRes)
	if err != nil {
		return err
	}

	mapOfAlphaAndIDs := make(map[float64]string)
	for _, l := range dl.Lines {
		mapOfAlphaAndIDs[l.Alpha] = ""
	}

	//find setup extgsate
	err = setupExtGStateToRes(ctx, dict, mapOfAlphaAndIDs)
	if err != nil {
		return err
	}
	//end setup extgsate

	//fmt.Printf("mapOfAlphaAndIDs: %+v\n", mapOfAlphaAndIDs)

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
		err := drawLinesToStream(&sm, dl.Lines, mapOfAlphaAndIDs)
		if err != nil {
			return err
		}
		entry.Object = sm
		//fmt.Printf("AAA\n")
	} else if ar, ok := obj.(pdfcpu.Array); ok {
		size := len(ar)
		if size <= 0 {
			return ErrDrawListNotSupport
		}
		o0 := ar[0]
		ir, _ := o0.(pdfcpu.IndirectRef)
		objNr = ir.ObjectNumber.Value()
		genNr := ir.GenerationNumber.Value()
		entry, _ := ctx.FindTableEntry(objNr, genNr)
		sm, _ := (entry.Object).(pdfcpu.StreamDict)
		err := drawLinesToStream(&sm, dl.Lines, mapOfAlphaAndIDs)
		if err != nil {
			return err
		}
		entry.Object = sm

		if size-1 > 0 {
			o1 := ar[size-1]
			ir, _ := o1.(pdfcpu.IndirectRef)
			objNr = ir.ObjectNumber.Value()
			genNr := ir.GenerationNumber.Value()
			entry, _ := ctx.FindTableEntry(objNr, genNr)
			sm, _ := (entry.Object).(pdfcpu.StreamDict)
			err := drawLinesToStream(&sm, dl.Lines, mapOfAlphaAndIDs)
			if err != nil {
				return err
			}
			entry.Object = sm
		}
		//fmt.Printf("BBB\n")
	} else {
		return ErrDrawListNotSupport
	}

	return nil
}

func drawLinesToStream(
	sd *pdfcpu.StreamDict,
	lines []Line,
	mapOfAlphaAndIDs map[float64]string,
) error {
	err := sd.Decode()
	if err != nil {
		return err
	}

	buff := bytes.NewBuffer(sd.Content)

	for _, line := range lines {

		writeStrokeColor(buff, line.Color, colorTypeStroke)

		l0 := fmt.Sprintf("%.2f w\n", line.LineWidth)
		l1 := fmt.Sprintf("%0.2f %0.2f m %0.2f %0.2f l S\n", line.X1, line.Y1, line.X2, line.Y2)
		buff.WriteString(gsString(mapOfAlphaAndIDs, line.Alpha))
		buff.WriteString(l0)
		buff.WriteString(l1)
	}
	sd.Content = buff.Bytes()
	//fmt.Printf("sd.Content: %s", string(sd.Content))
	return sd.Encode()
}

func gsString(mapOfAlphaAndIDs map[float64]string, alpha float64) string {
	for key, val := range mapOfAlphaAndIDs {
		if key == alpha {
			return fmt.Sprintf("/%s gs\n", val)
		}
	}
	return ""
}

func writeStrokeColor(buff *bytes.Buffer, rgb ColorRGB, colorType string) error {
	l := fmt.Sprintf("%.3f %.3f %.3f %s\n", float64(rgb.R)/255, float64(rgb.G)/255, float64(rgb.B)/255, colorType)
	_, err := buff.WriteString(l)
	if err != nil {
		return err
	}
	return nil
}

type ColorRGB struct {
	R int
	G int
	B int
}
