export GOPATH=$GOPATH:$(pwd)
# if [ -f ./core ]; then
#   rm ./core || true
# fi
go build -o ./shadow ./*.go;
echo "Build completed"
