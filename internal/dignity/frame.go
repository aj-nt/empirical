package dignity

// System identifies an astrological tradition or interpretive system.
type System string

const (
	SystemWestern System = "western"
	SystemKoiné   System = "koine"
	SystemVedic   System = "vedic"
)

// Frame identifies a coordinate reference frame for positions.
type Frame string

const (
	FrameTropical Frame = "tropical"
	FrameSidereal Frame = "sidereal"
	FrameDraconic Frame = "draconic"
	FrameCross    Frame = "cross"
)
