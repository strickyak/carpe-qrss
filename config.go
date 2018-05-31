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

	{"wa5djj", 80, "http://qsl.net/wa5djj/mf3.jpg"},
	{"wa5djj", 17, "http://qsl.net/wa5djj/hf5.jpg"},
	{"wa5djj", 15, "http://qsl.net/wa5djj/hf6.jpg"},
	{"wa5djj", 12, "http://qsl.net/wa5djj/hf7.jpg"},
	{"wa5djj", 10, "http://qsl.net/wa5djj/hf8.jpg"},

	{"wd4elg", 20, "https://dl.dropboxusercontent.com/s/gba72cz0au66032/WD4ELG%2020M%20grabber%20%28REFRESH%20for%20latest%20grab%29.jpg"},
	{"wd4elg", 30, "https://dl.dropboxusercontent.com/s/7djby65cbfh6hv7/WD4ELG%2030M%20grabber%20%28REFRESH%20for%20latest%20grab%29.jpg"},
	{"wd4elg", 40, "https://dl.dropboxusercontent.com/s/ajhc4t640k7k67u/WD4ELG%2040M%20grabber%20%28REFRESH%20for%20latest%20grab%29.jpg"},
	{"wd4elg", 80, "https://dl.dropboxusercontent.com/s/59ktcp48iie5i1m/WD4ELG%2080M%20grabber%20%28REFRESH%20for%20latest%20grab%29.jpg"},

	{"ve1vdm", 0, "http://users.eastlink.ca/~ve1vdm/argocaptures/argo.jpg"},
	{"zl2ik", 0, "http://zl2ik.com/Argo.jpg"},
	{"kl7l", 0, "http://kl7l.com/Alaska00000.jpg"},
}
