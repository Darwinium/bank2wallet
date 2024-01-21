# Bank2Wallet

Creates and generates a signed Apple Wallet Pass as a *.pkpass file that you can download and add to your Apple Wallet.

## Let's prepare everything

1. Create a folder `certificates` in the app folder.
2. Open Keychain App on your Mac, go to **Certificate Assistant** -> **Request Certificate from Authority**, write email and Save it in the `certificate` folder.
3. Go to https://developer.apple.com/account/resources/identifiers/list/passTypeId and create a new `Pass Type ID` by clicking **+** button. When prompted upload certificate created on the **step 2**.
4. Download passTypeId, double click and add it to Keychain.
5. Find `Pass Type ID` certificate under `My Certificates` (in login) and click Export. Don't forget the password used. Save the file in the `certificates` folder as `Certificate.p12`.
6. Open terminal and from the `certificates` folder run following commands:
   ```sh
   openssl pkcs12 -in Certificates.p12 -clcerts -nokeys -out passcertificate.pem -passin pass:<PASSWORD_YOU_USED_WHEN_EXPORTED> -legacy
   ```
   ```sh
   openssl pkcs12 -in Certificates.p12 -nocerts -out passkey.pem -passin pass: -passout pass:<PASSWORD_YOU_USED_WHEN_EXPORTED> -legacy
   ```
7. Don't change the names of the file in the cretificates folder
8. Create .env file:
   ```sh
   CERT_PASSWORD=<PASSWORD_YOU_USED_WHEN_EXPORTED>
   AUTH_TOKEN=<AUTH_TOKEN_THAT_YOU_GOT_FROM_FINOM>
   ```

## How to Run
To run this project, you can use Docker Compose:
```sh
docker-compose up
```

## Refferences and links
- Basic info about Apple Wallet Passes: https://developer.apple.com/documentation/walletpasses/creating_the_source_for_a_pass
- Design and patterns: https://developer.apple.com/library/archive/documentation/UserExperience/Conceptual/PassKit_PG/Creating.html#//apple_ref/doc/uid/TP40012195-CH4-SW1

