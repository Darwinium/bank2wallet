# Bank2Wallet

A secure and scalable Apple Wallet Pass generator for banking information. This application creates signed Apple Wallet passes (`.pkpass` files) that can be added to Apple Wallet, providing customers with quick access to their bank details including IBAN, BIC, cashback balance, and company information.

## Features

- **Apple Wallet Pass Generation**: Creates fully compliant Apple Wallet passes with banking information
- **Digital Signatures**: Signs passes with Apple certificates for authenticity
- **Real-time Updates**: Supports push notifications when pass information changes
- **QR Code Integration**: Includes EPC QR codes for SEPA payment compatibility
- **RESTful API**: Easy-to-use REST endpoints for pass management
- **Device Registration**: Manages device registrations for push notifications
- **Cashback Tracking**: Updates and notifies users of cashback balance changes
- **Database Persistence**: PostgreSQL database for reliable data storage

## Technology Stack

### Backend
- **Go 1.24** - Modern, performant backend language
- **Gin Framework** - Fast HTTP web framework
- **GORM** - ORM for database operations
- **PostgreSQL** - Reliable relational database
- **Zerolog** - Structured logging

### Infrastructure
- **Docker & Docker Compose** - Containerized deployment
- **OpenSSL** - Certificate management and pass signing
- **Apple Push Notification service (APNs)** - Push notification delivery

### Frontend
- **React 18.2** - Modern UI library
- **Axios** - HTTP client for API calls

## Architecture

```
bank2wallet/
├── server/              # Go backend application
│   ├── main.go         # HTTP server and routes
│   ├── db.go           # Database models and operations
│   ├── pass_generator.go  # Pass creation logic
│   ├── apn.go          # Push notification handling
│   ├── tools.go        # Utility functions
│   └── template/       # Pass design assets (images)
├── front/              # React frontend
├── db/                 # PostgreSQL configuration
└── docker-compose.yml  # Service orchestration
```

## Prerequisites

1. **macOS** (for certificate generation)
2. **Apple Developer Account** with Pass Type ID certificate
3. **Docker & Docker Compose**
4. **OpenSSL** (for certificate management)

## Setup Instructions

### 1. Apple Developer Certificate Setup

1. Create a `certificates` folder in the project root:
   ```bash
   mkdir certificates
   ```

2. Open **Keychain Access** on your Mac:
   - Go to **Certificate Assistant** → **Request a Certificate from a Certificate Authority**
   - Enter your email and save the certificate request in the `certificates` folder

3. Create a **Pass Type ID** on Apple Developer Portal:
   - Visit https://developer.apple.com/account/resources/identifiers/list/passTypeId
   - Click the **+** button to create a new Pass Type ID
   - Upload the certificate request created in step 2

4. Download and install the Pass Type ID certificate:
   - Download the certificate from Apple Developer Portal
   - Double-click to add it to Keychain

5. Export the certificate:
   - Find the **Pass Type ID** certificate under **My Certificates** in Keychain
   - Right-click and select **Export**
   - Save as `Certificates.p12` in the `certificates` folder
   - Set a password (you'll need this later)

6. Convert certificates to PEM format:
   ```bash
   cd certificates

   # Extract the certificate
   openssl pkcs12 -in Certificates.p12 -clcerts -nokeys -out passcertificate.pem -passin pass:<YOUR_PASSWORD> -legacy

   # Extract the private key
   openssl pkcs12 -in Certificates.p12 -nocerts -out passkey.pem -passin pass:<YOUR_PASSWORD> -passout pass:<YOUR_PASSWORD> -legacy
   ```

7. Download the Apple WWDR (Worldwide Developer Relations) certificate:
   ```bash
   # Download Apple's intermediate certificate
   wget https://www.apple.com/certificateauthority/AppleWWDRCAG4.cer -O AppleWWDRCAG4.cer
   openssl x509 -inform DER -outform PEM -in AppleWWDRCAG4.cer -out WWDR.pem
   ```

### 2. Environment Configuration

Create a `.env` file in the root directory:

```bash
# Application Settings
APP_ENV=release
GIN_MODE=release

# Server Configuration
SERVER_URL=0.0.0.0
SERVER_PORT=8080
WEB_SERVICE_URL=https://your-domain.com

# Certificate Password
CERT_PASSWORD=your_certificate_password

# Authentication Token
AUTH_TOKEN=your_secure_auth_token

# PostgreSQL Configuration
POSTGRES_HOST=db
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_db_password
POSTGRES_DB=bank2wallet

# PgAdmin (optional)
PGADMIN_EMAIL=admin@example.com
PGADMIN_PASSWORD=admin_password
```

### 3. Run with Docker Compose

```bash
# Start all services (database, server, pgadmin)
docker-compose up -d

# View logs
docker-compose logs -f server

# Stop all services
docker-compose down
```

The application will be available at:
- **API Server**: http://localhost:8080
- **Frontend**: http://localhost:3000 (if configured)
- **PgAdmin**: http://localhost:8888

## API Documentation

### Authentication

All API endpoints (except Apple Wallet specific endpoints) require an `Authorization` header:

```
Authorization: ApplePass <YOUR_AUTH_TOKEN>
```

### Endpoints

#### Create Pass
```http
POST /pass/v1/create
Content-Type: multipart/form-data

Parameters:
- companyID (required): Unique company identifier
- companyName (required): Company name
- iban (required): International Bank Account Number
- bic (required): Bank Identifier Code
- address (required): Company address
- cashback (optional): Cashback balance (default: 0)

Response:
{
  "message": "Pass was created successfully",
  "link": "https://your-domain.com/passes/uuid.pkpass",
  "companyID": "company123",
  "passID": "uuid"
}
```

#### Get Pass
```http
POST /pass/v1/getPass
Content-Type: multipart/form-data

Parameters:
- companyID (required): Company identifier

Response:
{
  "message": "Pass was retrieved successfully",
  "link": "https://your-domain.com/passes/uuid.pkpass",
  "companyID": "company123",
  "passID": "uuid"
}
```

#### Update Cashback
```http
POST /pass/v1/updateCashback
Content-Type: multipart/form-data

Parameters:
- companyID (required): Company identifier
- cashback (required): New cashback amount

Response:
{
  "message": "Cashback was updated successfully",
  "link": "https://your-domain.com/passes/uuid.pkpass",
  "companyID": "company123"
}
```

### Apple Wallet Integration Endpoints

These endpoints are called by Apple Wallet and the iOS system:

- `POST /pass/v1/registerDevice/v1/devices/:deviceLibraryIdentifier/registrations/:passTeamIdentifier/:serialNumber` - Register device for updates
- `GET /pass/v1/registerDevice/v1/devices/:deviceLibraryIdentifier/registrations/:passTeamIdentifier` - Check for pass updates
- `GET /pass/v1/registerDevice/v1/passes/:passTeamIdentifier/:serialNumber` - Get updated pass
- `DELETE /pass/v1/registerDevice/v1/devices/:deviceLibraryIdentifier/registrations/:passTeamIdentifier/:serialNumber` - Unregister device
- `POST /pass/v1/registerDevice/v1/log` - Receive logs from devices

## Development

### Local Development (without Docker)

1. **Start PostgreSQL**:
   ```bash
   docker run --name b2w-postgres \
     -e POSTGRES_PASSWORD=yourpassword \
     -p 5432:5432 -d postgres:latest
   ```

2. **Run the Go server**:
   ```bash
   cd server
   go mod download
   go run .
   ```

3. **Run the frontend** (optional):
   ```bash
   cd front
   npm install
   npm start
   ```

### Building

```bash
# Build Go application
cd server
go build -o bank2wallet .

# Build Docker image
docker build -t bank2wallet-server:latest ./server
```

### Testing

```bash
# Run tests
cd server
go test -v ./...

# Run tests with coverage
go test -v -cover ./...
```

## Database Schema

### Pass Table
- `id` (UUID, Primary Key): Unique pass identifier
- `company_id` (String): Company identifier
- `company_name` (String): Company name
- `iban` (String): International Bank Account Number
- `bic` (String): Bank Identifier Code
- `address` (String): Company address
- `cashback` (String): Cashback balance
- `created_at` (Timestamp): Creation time
- `updated_at` (Timestamp): Last update time

### Device Registration Table
- `device_library_identifier` (String): Device identifier
- `pass_type_identifier` (String): Pass type ID
- `serial_number` (String): Pass serial number
- `push_token` (String): APNs push token
- `created_at` (Timestamp): Registration time
- `updated_at` (Timestamp): Last update time

## Security Considerations

1. **Authentication**: All API endpoints are protected with token-based authentication
2. **HTTPS**: Always use HTTPS in production for the WEB_SERVICE_URL
3. **Secrets Management**: Never commit `.env` files or certificates to version control
4. **Token Security**: Use strong, randomly generated tokens for AUTH_TOKEN
5. **Database Security**: Use strong passwords and limit database access
6. **CORS**: Configure CORS appropriately for production (currently allows all origins)

## Troubleshooting

### Common Issues

1. **Certificate errors**: Ensure you're using the `-legacy` flag with OpenSSL for PKCS12 operations
2. **Database connection issues**: Verify PostgreSQL is running and credentials are correct
3. **Pass not downloading**: Check that WEB_SERVICE_URL is publicly accessible via HTTPS
4. **Push notifications not working**: Verify certificate password and APNs configuration

## References and Resources

- [Apple Wallet Pass Design Guidelines](https://developer.apple.com/library/archive/documentation/UserExperience/Conceptual/PassKit_PG/Creating.html)
- [PassKit Documentation](https://developer.apple.com/documentation/walletpasses)
- [EPC QR Code Specification](https://www.europeanpaymentscouncil.eu/document-library/guidance-documents/quick-response-code-guidelines-enable-data-capture-initiation)

## License

This project is proprietary software developed for Finom.

## Support

For issues and questions, please contact the development team or open an issue in the repository.
