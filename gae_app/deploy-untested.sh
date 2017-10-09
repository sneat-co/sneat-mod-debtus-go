#!/usr/bin/env bash

projectFolder=~/debtstracker
gaeAppFolder=${projectFolder}/gae_app/
deployFolder=~/debtstracker_deploy
testsBackupFolder=~/debtstracker_tests/

echo "Removing old files..."
rm -rf ${deployFolder}
echo "Copying new GAE files..."
cp -r ${gaeAppFolder}/appengine/ ${deployFolder}
echo "Removing test files..."
rm ${deployFolder}/*_test.go
#rsync --remove-source-files -a --prune-empty-dirs --include '*/' --include '*_test.go' --exclude '*' ${gaeAppFolder} ${testsBackupFolder}
echo "Copying new Ionic app files..."
cp -r ${projectFolder}/ionic-apps/public/platforms/browser/www/ ${deployFolder}/ionic-app
echo "//Cordova.js disabled in browser" > ${deployFolder}/ionic-app/cordova.js

while true; do
    read -p "Where do you want to deploy? (dev|prod): " app
    case $app in
        dev )
        	sed -i '' 's/^application: *[[:alpha:]]*/application: debtstracker-dev1/' ${deployFolder}/app.yaml
        	break;;
        prod )
        	sed -i '' 's/^application: *[[:alpha:]]*/application: debtstracker-io/' ${deployFolder}/app.yaml
        	break;;
        * ) echo "Please answer 'dev' or 'prod'.";;
    esac
done

echo "You selected: $app. File 'app.yaml' updated."

sed -i.bak 's/<script src="cordova.js"><\/script>/<!--script src="cordova.js"><\/script-->/' ${deployFolder}/ionic-app/index.html

#read -p "Check files:"

goapp deploy ${deployFolder}
#echo "Restoring test files..."
#rsync --remove-source-files -a --prune-empty-dirs --include '*/' --include '*_test.go' --exclude '*' ${testsBackupFolder} ${gaeAppFolder}
echo "DONE!"