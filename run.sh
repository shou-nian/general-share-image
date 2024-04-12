go mod download

go build github.com/aspecta-ai/look-share-img

# create _img temp directory
mkdir _img

chmod +x ./look-share-img
nohup ./look-share-img &
