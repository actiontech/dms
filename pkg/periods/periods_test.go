package periods

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPeriods_ScanValue(t *testing.T) {
	ps := Periods{
		&Period{
			StartHour:   1,
			StartMinute: 2,
			EndHour:     3,
			EndMinute:   4,
		},
		&Period{
			StartHour:   1,
			StartMinute: 3,
			EndHour:     2,
			EndMinute:   4,
		},
		&Period{
			StartHour:   2,
			StartMinute: 4,
			EndHour:     3,
			EndMinute:   1,
		},
	}

	data, err := ps.Value()
	assert.NoError(t, err)

	var ps2 Periods
	err = ps2.Scan(data)
	assert.NoError(t, err)
	assert.Equal(t, ps, ps2)

	ps3 := Periods{
		&Period{
			StartHour:   0,
			StartMinute: 2,
			EndHour:     3,
			EndMinute:   4,
		},
		&Period{
			StartHour:   0,
			StartMinute: 4,
			EndHour:     2,
			EndMinute:   4,
		},
		&Period{
			StartHour:   0,
			StartMinute: 4,
			EndHour:     3,
			EndMinute:   1,
		},
	}
	assert.NotEqual(t, ps3, ps2)

	var emptyPs Periods
	data, err = emptyPs.Value()
	assert.NoError(t, err)

	var emptyPs2 Periods
	err = emptyPs2.Scan(data)
	assert.NoError(t, err)
	assert.Equal(t, emptyPs, emptyPs2)

	data = []byte("this is test scan fail")
	var failPs Periods
	err = failPs.Scan(data)
	assert.Error(t, err)
}

func TestPeriods_SelfCheck(t *testing.T) {
	// Critical Values and Normal Intervals
	ps1 := Periods{
		{
			StartHour:   0,
			StartMinute: 0,
			EndHour:     23,
			EndMinute:   59,
		}, {
			StartHour:   1,
			StartMinute: 20,
			EndHour:     2,
			EndMinute:   10,
		},
	}
	assert.Equal(t, ps1.SelfCheck(), true)

	// The second rule end hour is earlier than start hour
	ps2 := Periods{
		{
			StartHour:   0,
			StartMinute: 0,
			EndHour:     23,
			EndMinute:   59,
		}, {
			StartHour:   2,
			StartMinute: 20,
			EndHour:     1,
			EndMinute:   10,
		},
	}
	assert.Equal(t, ps2.SelfCheck(), false)

	// The second rule end minute is equal than start minute
	ps3 := Periods{
		{
			StartHour:   0,
			StartMinute: 0,
			EndHour:     23,
			EndMinute:   59,
		}, {
			StartHour:   1,
			StartMinute: 1,
			EndHour:     1,
			EndMinute:   1,
		},
	}
	assert.Equal(t, ps3.SelfCheck(), false)

	// The second rule end minute is earlier than start minute
	ps4 := Periods{
		{
			StartHour:   0,
			StartMinute: 0,
			EndHour:     23,
			EndMinute:   59,
		}, {
			StartHour:   1,
			StartMinute: 10,
			EndHour:     1,
			EndMinute:   1,
		},
	}
	assert.Equal(t, ps4.SelfCheck(), false)

	// The second rule end hour is to large
	ps5 := Periods{
		{
			StartHour:   0,
			StartMinute: 0,
			EndHour:     23,
			EndMinute:   59,
		}, {
			StartHour:   1,
			StartMinute: 20,
			EndHour:     24,
			EndMinute:   10,
		},
	}
	assert.Equal(t, ps5.SelfCheck(), false)

	// The second rule end minutes is to large
	ps6 := Periods{
		{
			StartHour:   0,
			StartMinute: 0,
			EndHour:     23,
			EndMinute:   59,
		}, {
			StartHour:   1,
			StartMinute: 20,
			EndHour:     2,
			EndMinute:   60,
		},
	}
	assert.Equal(t, ps6.SelfCheck(), false)

	// The first start hour is too large
	ps7 := Periods{
		{
			StartHour:   24,
			StartMinute: 0,
			EndHour:     23,
			EndMinute:   59,
		}, {
			StartHour:   1,
			StartMinute: 20,
			EndHour:     2,
			EndMinute:   10,
		},
	}
	assert.Equal(t, ps7.SelfCheck(), false)

	//  The first start minute is too large
	ps8 := Periods{
		{
			StartHour:   0,
			StartMinute: 60,
			EndHour:     23,
			EndMinute:   59,
		}, {
			StartHour:   1,
			StartMinute: 20,
			EndHour:     2,
			EndMinute:   10,
		},
	}
	assert.Equal(t, ps8.SelfCheck(), false)

	//  The first start hour is too less
	ps9 := Periods{
		{
			StartHour:   -1,
			StartMinute: 0,
			EndHour:     23,
			EndMinute:   59,
		}, {
			StartHour:   1,
			StartMinute: 20,
			EndHour:     2,
			EndMinute:   10,
		},
	}
	assert.Equal(t, ps9.SelfCheck(), false)

	//  The first end minute is too less
	ps10 := Periods{
		{
			StartHour:   0,
			StartMinute: 0,
			EndHour:     23,
			EndMinute:   -4,
		}, {
			StartHour:   1,
			StartMinute: 20,
			EndHour:     2,
			EndMinute:   10,
		},
	}
	assert.Equal(t, ps10.SelfCheck(), false)

}

func TestPeriods_IsWithinScope(t *testing.T) {
	ps := Periods{
		{
			StartHour:   4,
			StartMinute: 3,
			EndHour:     5,
			EndMinute:   4,
		}, {
			StartHour:   2,
			StartMinute: 1,
			EndHour:     3,
			EndMinute:   2,
		},
	}

	// The first end threshold
	t0, err := time.Parse("2006-01-02 15:04:05", "2017-12-08 05:04:03")
	assert.NoError(t, err)
	assert.Equal(t, ps.IsWithinScope(t0), true)

	// The second start threshold
	t1, err := time.Parse("2006-01-02 15:04:05", "2017-12-08 02:01:03")
	assert.NoError(t, err)
	assert.Equal(t, ps.IsWithinScope(t1), true)

	// in the first interval
	t2, err := time.Parse("2006-01-02 15:04:05", "2017-12-08 03:01:53")
	assert.NoError(t, err)
	assert.Equal(t, ps.IsWithinScope(t2), true)

	// too early
	t3, err := time.Parse("2006-01-02 15:04:05", "2017-12-08 01:01:53")
	assert.NoError(t, err)
	assert.Equal(t, ps.IsWithinScope(t3), false)

	// between two periods
	t4, err := time.Parse("2006-01-02 15:04:05", "2017-12-08 03:03:53")
	assert.NoError(t, err)
	assert.Equal(t, ps.IsWithinScope(t4), false)

	// too late
	t5, err := time.Parse("2006-01-02 15:04:05", "2017-12-08 23:01:53")
	assert.NoError(t, err)
	assert.Equal(t, ps.IsWithinScope(t5), false)

}

func TestParsePeriods(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    Periods
		wantErr bool
	}{
		{
			name:    "blank string",
			args:    "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid string",
			args:    "whatever",
			want:    nil,
			wantErr: true,
		},

		{
			name: "ok 1 period",
			args: "09:30-11:30",
			want: []*Period{
				{9, 30, 11, 30},
			},
			wantErr: false,
		},
		{
			name: "ok 1 period, support(-0)",
			args: "-0:10-11:-0",
			want: []*Period{
				{0, 10, 11, 0},
			},
			wantErr: false,
		},
		{
			name:    "fail 1 period",
			args:    "09:30a-11:30",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "fail 1 period",
			args:    "09:30 11:30",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "fail 1 period",
			args:    "09:00-11:-10",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "fail 1 period, invalid minute",
			args:    "09:60-11:30",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "fail 1 period, invalid hour",
			args:    "09:30-24:30",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "fail 1 period",
			args:    "9:30-11:300",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "fail 1 period, invalid period",
			args:    "09:30-08:30",
			want:    nil,
			wantErr: true,
		},

		{
			name: "ok 2 periods, no leading zero",
			args: "9:30-11:30;11:30-13:30",
			want: []*Period{
				{9, 30, 11, 30},
				{11, 30, 13, 30},
			},
			wantErr: false,
		},
		{
			name: "ok 2 periods, periods disorder",
			args: "11:30-13:30;9:30-11:30",
			want: []*Period{
				{11, 30, 13, 30},
				{9, 30, 11, 30},
			},
			wantErr: false,
		},
		{
			name: "ok 2 periods,support(-0)",
			args: "-0:-0--0:30;9:-0-11:-0",
			want: []*Period{
				{0, 0, 0, 30},
				{9, 0, 11, 0},
			},
			wantErr: false,
		},
		{
			name:    "fail 2 periods,blank",
			args:    "11:30-13:30;",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "fail 2 periods, unexpected newline",
			args:    "11:3-13:3;21:0-\n23:0",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "fail 2 periods, periods overlap",
			args:    "9:3-51:30;9:30-21:30",
			want:    nil,
			wantErr: true,
		},

		{
			name: "ok 3 periods disorder",
			args: "09:30-11:30;20:30-21:30;11:30-13:30",
			want: []*Period{
				{9, 30, 11, 30},
				{20, 30, 21, 30},
				{11, 30, 13, 30},
			},
			wantErr: false,
		},
		{
			name: "ok 3 periods, no leading zero",
			args: "9:3-11:30;20:30-21:30;11:30-13:30",
			want: []*Period{
				{9, 3, 11, 30},
				{20, 30, 21, 30},
				{11, 30, 13, 30},
			},
			wantErr: false,
		},
		{
			name:    "fail 3 periods, periods overlap",
			args:    "9:3-51:30;9:30-21:30;11:30-13:30",
			want:    nil,
			wantErr: true,
		},

		{
			name: "ok",
			args: "01:30-2:00;2:30-3:0;6:30-7:0;3:30-4:0;7:30-8:0;8:30-9:00;9:30-10:0",
			want: []*Period{
				{1, 30, 2, 0},
				{2, 30, 3, 0},
				{6, 30, 7, 0},
				{3, 30, 4, 0},
				{7, 30, 8, 0},
				{8, 30, 9, 0},
				{9, 30, 10, 0},
			},
			wantErr: false,
		},
		{
			name:    "fail ;;",
			args:    "01:30-2:00;2:30-3:0;;6:30-7:0;3:30-4:0;7:30-8:0;8:30-9:00;9:30-10:0",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePeriods(tt.args)
			if err != nil {
				t.Logf("ParsePeriods result: %v\n", err)
			}
			assert.Equalf(t, tt.wantErr, err != nil, "ParsePeriods(%v)", tt.args)
			assert.Equalf(t, tt.want, got, "ParsePeriods(%v)", tt.args)
		})
	}
}
