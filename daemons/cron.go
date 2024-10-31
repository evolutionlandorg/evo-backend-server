package daemons

import "context"

func StartUploadData(ctx context.Context) {
	UploadProjectData(ctx, "heco")
}
