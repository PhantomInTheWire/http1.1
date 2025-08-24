package request

const (
	supportedHttpVersion = "1.1"
)

const (
	requestLineMethodIndex   = 0
	requestLineTargetIndex   = 1
	requestLineVersionIndex  = 2
	expectedRequestLineParts = 3
)

const (
	httpVersionPartsCount = 2
	httpVersionValueIndex = 1
)

const (
	requestLineDataIndex = 0
	minDataPartsCount    = 1
)

const (
	emptyString            = ""
	spaceDelimiter         = " "
	carriageReturnLineFeed = "\r\n"
	slashDelimiter         = "/"
)
