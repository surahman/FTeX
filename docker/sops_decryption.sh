#!/bin/bash

configFilePath=$1
encryptExt=sops
decryptExt=yaml

###############################################################################
#               Decrypt Secret OperationS Configuration Files                 #
###############################################################################

echo 'Decrypting Secret OperationS configuration files...'

for item in "$configFilePath"/*."$encryptExt"; do
  filename=$(basename "$item" ."$encryptExt")

  sops -d --input-type yaml --output-type yaml "$item" > "$configFilePath"/"$filename"."$decryptExt" ||
  { echo "Decrypting $item failed"; exit 1; }

  rm "$item"
done

echo 'Decryption completed successfully.'
