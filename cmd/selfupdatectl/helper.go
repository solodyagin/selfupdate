package main

func buildS3Path(baseS3Path string, exe string) string {
	s3path := ""
	if baseS3Path != "" {
		s3path = baseS3Path
		if baseS3Path[len(baseS3Path)-1] != '/' {
			s3path += "/"
		}
	}
	s3path += exe

	return s3path
}
