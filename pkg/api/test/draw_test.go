package test

import (
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func TestDrawLineFile(t *testing.T) {

	fileName := "Acroforms2.pdf"
	src := filepath.Join(inDir, fileName)
	//src := "/Users/oneplus/Code/Work/gopdf/test/out/result1_by_parsed_ttf_font.pdf"
	//src := "/Users/oneplus/testxx/line.pdf"
	//src := "/Users/oneplus/Code/Work/gopdf/test/out/line2.pdf"
	//src := "/Users/oneplus/Code/Work/ebooks-server/testing/pdf/ff.pdf"
	dest := filepath.Join(outDir, fileName)
	//dest := "/Users/oneplus/Code/Work/gopdf/test/out/result1_by_parsed_ttf_font_out.pdf"
	//dest := "/Users/oneplus/testxx/out/line.pdf"
	//dest := "/Users/oneplus/Code/Work/gopdf/test/out/ff_out.pdf"
	//dest := "/Users/oneplus/Code/Work/gopdf/test/out/Acroforms2.pdf"

	draws := []api.DrawLine{
		{
			PageNumber: 1,
			Lines: []api.Line{
				{
					Alpha: 0.9,
					Color: api.ColorRGB{
						R: 255,
						G: 0,
						B: 0,
					},
					LineWidth: 2.0,
					X1:        0,
					Y1:        0,
					X2:        800,
					Y2:        800,
				}, {
					Alpha: 0.2,
					Color: api.ColorRGB{
						R: 0,
						G: 0,
						B: 255,
					},
					LineWidth: 20.0,
					X1:        100,
					Y1:        10,
					X2:        800,
					Y2:        800,
				},
			},
		},
	}

	err := api.DrawLineFile(src, dest, draws, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
}
