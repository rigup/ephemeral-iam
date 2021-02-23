package eiamutil

// CheckError handles simple error handling
func CheckError(err error) {
	if err != nil {
		Logger.Fatalln(err)
	}
}
