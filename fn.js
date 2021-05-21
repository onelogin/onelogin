hello there

const axios = require("axios");
const jwt = require('jsonwebtoken');
const jwksClient = require('jwks-rsa');

function getKey(header, callback){
  var client = jwksClient({
    jwksUri: "https://" + process.env.AUTH0_SUBDOMAIN + ".auth0.com/.well-known/jwks.json"
  });

  client.getSigningKey(header.kid, function(err, key) {
    var signingKey = key.publicKey || key.rsaPublicKey;
    callback(null, signingKey);
  });
}

async function verifyJWT(token) {
  return new Promise(
    (resolve, reject) => {
      jwt.verify(token, getKey, {}, function(err, decoded) {
        resolve(decoded)
      });      
    }
  );
}

exports.handler = async (context) => {
    // Its not considered good practice to log the user context on 
    // this hook as it contains a password. Only enable this for debugging
    // console.log(context);

    let user;

    try {
      // Do a password grant request to validate the password
      const response = await axios.post("https://" + process.env.AUTH0_SUBDOMAIN + ".auth0.com/oauth/token", {
          grant_type: "password",
          username: context.user_identifier,
          password: context.password,
          scope: "openid",
          client_id: process.env.AUTH0_CLIENT_ID,
          client_secret: process.env.AUTH0_CLIENT_SECRET
      }, {
          headers: { 
              "Content-Type": "application/json"
          }
      });
      // console.log(response);

      if (response.data) {
        let decodedToken = await verifyJWT(response.data.id_token);
        console.log(decodedToken);

        let name = decodedToken.name.split(" ");

        return {
          success: true,
          user: {
            username: context.user_identifier,
            password: context.password,
            firstname: name.shift(),
            lastname: name.join(" "),
            email: decodedToken.email
          }
        };          
      }
    }
    catch (error) {
        console.log("Error authenticating user ", error);         
    }  
    
    // Fail closed. Dont create the user. Deny access
    return {
        success: false,
        user: {}
    }       
}
