#!/bin/bash
if [[ -z $1 ]]
then
    echo "please provide Package path as argument to the script"
    exit 1
fi

if [[ -z $2 ]]
then
    echo "please provide service name as argument to the script"
    exit 1
fi

PKG=$1

SVC=$2
SVC=`echo ${SVC:0:1} | tr  '[a-z]' '[A-Z]'`${SVC:1}

echo "creating $SVC"
#copy to servicename
SVC_PATH=$GOPATH/src/$PKG/$SVC/
mkdir -p $SVC_PATH
cp -r . $SVC_PATH
rm -rf $SVC_PATH/.git
rm -rf $SVC_PATH/create.sh
frame=`pwd`
cd $SVC_PATH; git init

echo "Updating source code"
# Update Code/Config
for file in $(grep -rl "ServiceName" * | grep -v vendor| grep -v rename)
do
    # don't replace lines starting with newrelicServiceName, else do
    sed -i "" "/^newrelicServiceName/!s/ServiceName/$SVC/g" $file
done

rm ServiceName/ServiceName_proto/*.pb.go

echo "Updating project structure"

#Update Folders
mv ServiceName $SVC
for file in $(find . -maxdepth 10 -type d -name '*ServiceName*' -print)
do
    mv $file ${file/ServiceName/$SVC}
done

#update files
for file in $(find . -maxdepth 10 -type f -name '*ServiceName*' -print)
do
    mv $file ${file/ServiceName/$SVC}
done

echo "creating $SVC project"
# remove builder
for file in $(grep -rl "Orion/builder" * | grep -v vendor| grep -v rename)
do
    echo $file
    sed -i "" "s|github.com/carousell/Orion/builder|$PKG/Orion/builder|g" $file
    sed -i "" "s|Orion/builder|$SVC|g" $file
done

echo "regenerating protobuf"
p=`pwd`
cd ${SVC}/${SVC}_proto/
bash generate.sh
cd $p

svc_name=`echo "$SVC" | awk '{print tolower($0)}'`
sed -i "" "s/service_name/$SVC/g" ./sonar.properties
sed -i "" "s/service_name/$svc_name/g" run.sh

echo "compiling code for $SVC"
make build
git add .
git commit -m "$SVC created from https://github.com/carousell/Orion"
echo ""
echo "$SVC initialized in "`pwd`
cd $frame
