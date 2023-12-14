package biz

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/agiledragon/gomonkey"
	"github.com/smartystreets/goconvey/convey"
)

func TestLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	fmt.Println("test")

	convey.Convey("Given some integer with a starting value", t, func() {
		x := 1

		convey.Convey("When the integer is incremented", func() {
			x++

			convey.Convey("The value should be greater by one", func() {
				convey.So(x, convey.ShouldEqual, 2)
			})
		})
	})

	s := &UserUsecase{}
	patches := gomonkey.ApplyMethod(reflect.TypeOf(s), "NewDoer", func(s *UserUsecase) {
	})
	defer patches.Reset()
}
