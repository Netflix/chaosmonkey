package spinnaker

func Fuzz(data []byte) int {
	_, err := fromJSON(data)
	if err != nil {
		return 0
	}
	return 1
}
