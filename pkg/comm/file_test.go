package comm

import (
	"reflect"
	"testing"

	"github.com/huantedness/autowire/pkg/util"
)

func Test_renamed(t *testing.T) {
	type args struct {
		bm   util.BiDirectionMap[path, alias]
		path path
		name alias
	}
	tests := []struct {
		name string
		args args
		want alias
	}{
		{
			name: "0 conflict exist, first time population",
			args: args{
				bm: func() util.BiDirectionMap[path, alias] {
					bm := util.NewBiDirectionMap[path, alias]()
					return bm
				}(),
				path: "github.com/hauntedness/autowire/pkg",
				name: "pkg",
			},
			want: "pkg",
		},
		{
			name: "normal package path, 0 conflict exist",
			args: args{
				bm: func() util.BiDirectionMap[path, alias] {
					bm := util.NewBiDirectionMap[path, alias]()
					bm.MustPut("github.com/hauntedness/autowire/pkg", "pkg")
					return bm
				}(),
				path: "github.com/hauntedness/autowire/pkg",
				name: "pkg",
			},
			want: "pkg",
		},
		{
			name: "1 conflict exists, module name doesn't contain strange charactors",
			args: args{
				bm: func() util.BiDirectionMap[path, alias] {
					bm := util.NewBiDirectionMap[path, alias]()
					bm.MustPut("github.com/hauntedness/autowire/pkg", "pkg")
					return bm
				}(),
				path: "github.com/hauntedness/others/pkg",
				name: "pkg",
			},
			want: "otherspkg",
		},
		{
			name: "1 conflict exists, but package already renamed",
			args: args{
				bm: func() util.BiDirectionMap[path, alias] {
					bm := util.NewBiDirectionMap[path, alias]()
					bm.MustPut("github.com/hauntedness/autowire/pkg", "pkg")
					bm.MustPut("github.com/hauntedness/others/pkg", "otherspkg")
					return bm
				}(),
				path: "github.com/hauntedness/others/pkg",
				name: "pkg",
			},
			want: "otherspkg",
		},
		{
			name: "2 conflicts exist, module name(or parent package path) contain strange charactors",
			args: args{
				bm: func() util.BiDirectionMap[path, alias] {
					bm := util.NewBiDirectionMap[path, alias]()
					bm.MustPut("github.com/hauntedness/autowire/pkg", "pkg")
					bm.MustPut("github.com/someotheruser1/1some$sign_here/pkg", "_some_sign_herepkg")
					return bm
				}(),
				path: "github.com/someotheruser2/1some$sign_here/pkg",
				name: "pkg",
			},
			want: "_some_sign_herepkg1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renamed(tt.args.bm, tt.args.path, tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("renamed() = %v, want %v", got, tt.want)
			}
		})
	}
}
