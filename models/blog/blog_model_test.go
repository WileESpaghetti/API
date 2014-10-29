package blog_model

import (
	. "github.com/smartystreets/goconvey/convey"
	"math/rand"
	"testing"
	"time"
)

func TestGetBlogs(t *testing.T) {
	Convey("Testing Gets", t, func() {
		Convey("Testing GetAll()", func() {
			var bs Blogs
			var err error
			bs, err = GetAll()
			So(bs, ShouldHaveSameTypeAs, Blogs{})
			So(err, ShouldBeNil)
			So(len(bs), ShouldNotBeNil)

			if len(bs) > 0 {
				x := rand.Intn(len(bs))
				Convey("Testing Get()", func() {
					b := Blog{
						ID: bs[x].ID,
					}
					err = b.Get()
					So(err, ShouldBeNil)
					So(b.Title, ShouldNotEqual, "")
					So(b.Slug, ShouldNotEqual, "")
					So(b.PublishedDate, ShouldHaveSameTypeAs, time.Time{})

					b = Blog{
						ID: bs[len(bs)-1].ID + 1,
					}
					err = b.Get()
					So(err, ShouldBeNil)
					So(b.Title, ShouldEqual, "")
				})

			}
		})

		Convey("Testing GetAllCategories()", func() {
			qs, err := GetAllCategories()
			So(qs, ShouldHaveSameTypeAs, Categories{})
			So(err, ShouldBeNil)
		})
		Convey("Testing Search()", func() {
			as, err := Search("test", "", "", "", "", "", "", "", "", "", "", "1", "0")
			So(as, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(as.Pagination.Page, ShouldEqual, 1)
			So(as.Pagination.ReturnedCount, ShouldNotBeNil)
			So(as.Pagination.PerPage, ShouldNotBeNil)
			So(as.Pagination.PerPage, ShouldEqual, len(as.Objects))
		})

	})
	Convey("Testing CUD", t, func() {
		Convey("Testing Create()/Delete()", func() {
			var f Blog
			var cs BlogCategories
			var c BlogCategory
			var err error
			f.Title = "testTitle"
			f.Slug = "testSlug"
			f.Text = "test"
			f.PublishedDate, err = time.Parse(timeFormat, "2004-03-03 9:15:00")
			f.UserID = 1
			f.MetaTitle = "test"
			f.MetaDescription = "test"
			f.Keywords = "test"
			f.Active = true
			c.Category.Active = true
			c.Category.Name = "testCat"
			c.Category.Slug = "catSlug"
			cs = append(cs, c)
			f.BlogCategories = cs

			f.Create()
			f.Get()
			So(f, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(f.Title, ShouldEqual, "testTitle")
			So(f.Slug, ShouldEqual, "testSlug")
			var t time.Time
			So(f.PublishedDate, ShouldHaveSameTypeAs, t)

			f.Title = "testTitle222"
			f.Slug = "testSlug222"
			f.PublishedDate, err = time.Parse(timeFormat, "2004-03-03 09:15:00")

			ch := make(chan int)
			go func() {
				f.Update()
				ch <- 1
			}()
			<-ch
			f.Get()
			So(err, ShouldBeNil)
			So(f, ShouldNotBeNil)
			So(f.Title, ShouldEqual, "testTitle222")
			So(f.Slug, ShouldEqual, "testSlug222")

			err = f.Delete()
			So(err, ShouldBeNil)
		})

		Convey("Testing CreateCategory()/DeleteCategory()", func() {
			var c Category
			var err error
			c.Name = "testTitle"
			c.Slug = "testSlug"
			c.Active = true
			c.Create()
			c.Get()
			So(c, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(c.Name, ShouldEqual, "testTitle")
			So(c.Slug, ShouldEqual, "testSlug")
			So(c.Active, ShouldBeTrue)
			err = c.Delete()
			So(err, ShouldBeNil)

		})

	})

}
