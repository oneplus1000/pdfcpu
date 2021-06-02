package test

import (
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func TestDrawLineFile(t *testing.T) {

	//fileName := "Acroforms2.pdf"
	//src := filepath.Join(inDir, fileName)
	//src := "/Users/oneplus/Code/Work/gopdf/test/out/result1_by_parsed_ttf_font.pdf"
	//src := "/Users/oneplus/Code/Work/gopdf/test/out/number_of_pages_test.pdf"
	src := "/Users/oneplus/Code/Work/ebooks-server/testing/pdf/ff.pdf"
	//dest := filepath.Join(outDir, fileName)
	//dest := "/Users/oneplus/Code/Work/gopdf/test/out/result1_by_parsed_ttf_font_out.pdf"
	//dest := "/Users/oneplus/Code/Work/gopdf/test/out/number_of_pages_test_out.pdf"
	dest := "/Users/oneplus/Code/Work/gopdf/test/out/ff_out.pdf"

	draws := []api.DrawLine{
		{
			PageNumber: 1,
			Lines: []api.Line{
				{
					LineWidth: 2.0,
					X1:        0,
					Y1:        0,
					X2:        800,
					Y2:        800,
				}, {
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
