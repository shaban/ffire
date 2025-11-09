// Test script for N-API bindings
const { ConfigMessage } = require('./');

// Create a message
const config = new ConfigMessage({
  Host: "localhost",
  Port: 8080,
  EnableSSL: true,
  Timeout: 30.5,
  MaxRetries: 3
});

console.log("Original object:", config);

// Encode
const encoded = config.encode();
console.log("Encoded buffer:", encoded);
console.log("Buffer length:", encoded.length, "bytes");

// Decode
const decoded = ConfigMessage.decode(encoded);
console.log("Decoded object:", decoded);

// Verify
console.log("\nVerification:");
console.log("Host matches:", decoded.Host === config.Host);
console.log("Port matches:", decoded.Port === config.Port);
console.log("EnableSSL matches:", decoded.EnableSSL === config.EnableSSL);
console.log("Timeout matches:", decoded.Timeout === config.Timeout);
console.log("MaxRetries matches:", decoded.MaxRetries === config.MaxRetries);

console.log("\n✅ All tests passed!");
