package crawl_test

import "testing"

func Test_Crawl_single_file_input(t *testing.T) {
	ttests := map[string]struct {
		objType any
	}{
		"test1": {
			objType: nil,
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			_ = tt
		})
	}
}
