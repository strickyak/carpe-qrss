package carpe

type Target struct {
	Nick string
	Band int
	URL  string
}

var Targets = []Target{
	{"w4hbk", 30, "http://www.qsl.net/w4hbk/SL1.jpg"},
	{"w4hbk-4hr", 30, "http://www.qsl.net/w4hbk/4Hrs.jpg"},

	{"wa5djj", 40, "http://qsl.net/wa5djj/hf2.jpg"},
	{"wa5djj", 30, "http://qsl.net/wa5djj/hf3.jpg"},
	{"wa5djj", 20, "http://qsl.net/wa5djj/hf4.jpg"},

	// {"wa5djj", 80, "http://qsl.net/wa5djj/mf3.jpg"},
	// {"wa5djj", 17, "http://qsl.net/wa5djj/hf5.jpg"},
	// {"wa5djj", 15, "http://qsl.net/wa5djj/hf6.jpg"},
	// {"wa5djj", 12, "http://qsl.net/wa5djj/hf7.jpg"},
	// {"wa5djj", 10, "http://qsl.net/wa5djj/hf8.jpg"},

	{"wd4elg", 20, "https://dl.dropboxusercontent.com/s/gba72cz0au66032/WD4ELG%2020M%20grabber%20%28REFRESH%20for%20latest%20grab%29.jpg"},
	{"wd4elg", 30, "https://dl.dropboxusercontent.com/s/7djby65cbfh6hv7/WD4ELG%2030M%20grabber%20%28REFRESH%20for%20latest%20grab%29.jpg"},
	{"wd4elg", 40, "https://dl.dropboxusercontent.com/s/ajhc4t640k7k67u/WD4ELG%2040M%20grabber%20%28REFRESH%20for%20latest%20grab%29.jpg"},
	{"wd4elg", 80, "https://dl.dropboxusercontent.com/s/59ktcp48iie5i1m/WD4ELG%2080M%20grabber%20%28REFRESH%20for%20latest%20grab%29.jpg"},

	{"ve1vdm", 0, "http://users.eastlink.ca/~ve1vdm/argocaptures/argo.jpg"},
	{"zl2ik", 0, "http://zl2ik.com/Argo.jpg"},
	{"kl7l", 0, "http://kl7l.com/Alaska00000.jpg"},
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
