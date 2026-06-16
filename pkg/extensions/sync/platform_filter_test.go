//go:build sync

package sync

import (
	"testing"

	ispec "github.com/opencontainers/image-spec/specs-go/v1"
	. "github.com/smartystreets/goconvey/convey"
)

func TestParsePlatformSpec(t *testing.T) {
	Convey("parsePlatformSpec", t, func() {
		Convey("arch only", func() {
			p := parsePlatformSpec("amd64")
			So(p.arch, ShouldEqual, "amd64")
			So(p.os, ShouldBeEmpty)
			So(p.variant, ShouldBeEmpty)
		})

		Convey("os/arch", func() {
			p := parsePlatformSpec("linux/amd64")
			So(p.os, ShouldEqual, "linux")
			So(p.arch, ShouldEqual, "amd64")
			So(p.variant, ShouldBeEmpty)
		})

		Convey("os/arch/variant", func() {
			p := parsePlatformSpec("linux/arm/v7")
			So(p.os, ShouldEqual, "linux")
			So(p.arch, ShouldEqual, "arm")
			So(p.variant, ShouldEqual, "v7")
		})
	})
}

func TestMatchesPlatform(t *testing.T) {
	Convey("matchesPlatform", t, func() {
		Convey("empty specs matches everything", func() {
			p := &ispec.Platform{OS: "linux", Architecture: "amd64"}
			So(matchesPlatform(p, nil), ShouldBeTrue)
			So(matchesPlatform(p, []string{}), ShouldBeTrue)
		})

		Convey("nil platform matches everything", func() {
			So(matchesPlatform(nil, []string{"linux/amd64"}), ShouldBeTrue)
		})

		Convey("os/arch match", func() {
			p := &ispec.Platform{OS: "linux", Architecture: "amd64"}
			So(matchesPlatform(p, []string{"linux/amd64"}), ShouldBeTrue)
			So(matchesPlatform(p, []string{"linux/arm64"}), ShouldBeFalse)
		})

		Convey("arch-only match", func() {
			p := &ispec.Platform{OS: "linux", Architecture: "amd64"}
			So(matchesPlatform(p, []string{"amd64"}), ShouldBeTrue)
		})

		Convey("os/arch/variant match", func() {
			p := &ispec.Platform{OS: "linux", Architecture: "arm", Variant: "v7"}
			So(matchesPlatform(p, []string{"linux/arm/v7"}), ShouldBeTrue)
			So(matchesPlatform(p, []string{"linux/arm/v6"}), ShouldBeFalse)
		})

		Convey("variant in spec but not in platform still matches", func() {
			p := &ispec.Platform{OS: "linux", Architecture: "arm"}
			So(matchesPlatform(p, []string{"linux/arm/v7"}), ShouldBeTrue)
		})

		Convey("multiple specs, one matches", func() {
			p := &ispec.Platform{OS: "linux", Architecture: "arm64"}
			So(matchesPlatform(p, []string{"linux/amd64", "linux/arm64"}), ShouldBeTrue)
		})

		Convey("multiple specs, none match", func() {
			p := &ispec.Platform{OS: "windows", Architecture: "amd64"}
			So(matchesPlatform(p, []string{"linux/amd64", "linux/arm64"}), ShouldBeFalse)
		})

		Convey("wrong os", func() {
			p := &ispec.Platform{OS: "windows", Architecture: "amd64"}
			So(matchesPlatform(p, []string{"linux/amd64"}), ShouldBeFalse)
		})
	})
}

func TestFormatPlatform(t *testing.T) {
	Convey("formatPlatform", t, func() {
		Convey("nil", func() {
			So(formatPlatform(nil), ShouldEqual, "unknown")
		})

		Convey("os/arch", func() {
			p := &ispec.Platform{OS: "linux", Architecture: "amd64"}
			So(formatPlatform(p), ShouldEqual, "linux/amd64")
		})

		Convey("os/arch/variant", func() {
			p := &ispec.Platform{OS: "linux", Architecture: "arm", Variant: "v7"}
			So(formatPlatform(p), ShouldEqual, "linux/arm/v7")
		})
	})
}
