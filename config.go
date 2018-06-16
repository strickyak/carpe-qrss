package carpe

type Target struct {
	Nick string
	URL  string
}

var Targets = []Target{
	{"w4hbk", "http://www.qsl.net/w4hbk/SL1.jpg"},
	{"w4hbk_4h", "http://www.qsl.net/w4hbk/4Hrs.jpg"},

	{"wa5djj_40", "http://qsl.net/wa5djj/hf2.jpg"},
	{"wa5djj_30", "http://qsl.net/wa5djj/hf3.jpg"},
	{"wa5djj_20", "http://qsl.net/wa5djj/hf4.jpg"},

	{"wa5djj_80", "http://qsl.net/wa5djj/mf3.jpg"},
	{"wa5djj_17", "http://qsl.net/wa5djj/hf5.jpg"},
	{"wa5djj_15", "http://qsl.net/wa5djj/hf6.jpg"},
	{"wa5djj_12", "http://qsl.net/wa5djj/hf7.jpg"},
	{"wa5djj_10", "http://qsl.net/wa5djj/hf8.jpg"},

	// {"wd4elg_20", "https://dl.dropboxusercontent.com/s/gba72cz0au66032/WD4ELG%2020M%20grabber%20%28REFRESH%20for%20latest%20grab%29.jpg"},
	// {"wd4elg_30", "https://dl.dropboxusercontent.com/s/7djby65cbfh6hv7/WD4ELG%2030M%20grabber%20%28REFRESH%20for%20latest%20grab%29.jpg"},
	// {"wd4elg_40", "https://dl.dropboxusercontent.com/s/ajhc4t640k7k67u/WD4ELG%2040M%20grabber%20%28REFRESH%20for%20latest%20grab%29.jpg"},
	// {"wd4elg_80", "https://dl.dropboxusercontent.com/s/59ktcp48iie5i1m/WD4ELG%2080M%20grabber%20%28REFRESH%20for%20latest%20grab%29.jpg"},

	{"ve1vdm", "http://users.eastlink.ca/~ve1vdm/argocaptures/argo.jpg"},
	{"zl2ik", "http://zl2ik.com/Argo.jpg"},
	{"kl7l", "http://kl7l.com/Alaska00000.jpg"},

	// Part-time 630m grabber:
	{"tg6ajr", "http://qsl.net/tg9ajr/argo/mf1.gif"},

	// Holywell, Northumberland, IO95FB
	{"g3vyz_1", "http://www.holywell44.com/qrss/qrss_.jpg"},
	{"g3vyz_2", "http://www.holywell44.com/qrss/qrss_2.jpg"},
	{"g3vyz_3", "http://www.holywell44.com/qrss/qrss_3.jpg"},
	{"g3vyz_4", "http://www.holywell44.com/qrss/qrss_4.jpg"},
	{"g3vyz_5", "http://www.holywell44.com/qrss/qrss_5.jpg"},

	// Gran Canaria Island (locator IL28fd)
	{"ea8bvp", "http://www.qsl.net/ea8bvp/hf1.jpg"},
	{"ea8bvp_4h", "http://www.qsl.net/ea8bvp/hf2.jpg"},

	{"la5goa", "http://la5goa.manglet.net/grabber/lopshot1.jpg"},

	{"n2nxz", "http://www.qsl.net/n2nxz/hf1.jpg"},

	{"wd4ah", "http://www.qsl.net/wd4ah/hf1.jpg"},

	{"ok1fcx", "http://qsl.net/ok1fcx/hf1.jpg"},
}

type CropMargins []int
type OriginalDim int // encodes (width<<16 + height)

func MakeOriginalDim(width, height int) OriginalDim {
	return OriginalDim(width<<16 + height)
}

// Map original (width, height) to (left, right, top, bottom) margins.
// So far, each different type of image has a different size.
// We will use that, as long as it works.
var Crops = map[OriginalDim]CropMargins{
	MakeOriginalDim(1152, 702): []int{0, 80, 0, 80},   // wa5djj
	MakeOriginalDim(1000, 696): []int{12, 96, 88, 40}, // kl7l
	MakeOriginalDim(1235, 686): []int{3, 130, 5, 5},   // kl7l
	MakeOriginalDim(1226, 721): []int{4, 172, 4, 4},   // ve1vdm
	MakeOriginalDim(1187, 812): []int{3, 262, 3, 3},   // w4hbk
}
