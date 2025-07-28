package com.hivestudent.bookme.services;

import io.jsonwebtoken.Jwts;
import lombok.RequiredArgsConstructor;
import lombok.SneakyThrows;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.ParameterizedTypeReference;
import org.springframework.http.MediaType;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestClient;

import java.nio.charset.StandardCharsets;
import java.security.KeyFactory;
import java.security.NoSuchAlgorithmException;
import java.security.PrivateKey;
import java.security.spec.InvalidKeySpecException;
import java.security.spec.PKCS8EncodedKeySpec;
import java.util.Base64;
import java.util.Date;

import static java.net.URLEncoder.encode;

@Service
@RequiredArgsConstructor
public class GoogleService {

    private final RestClient restClient;

    @Value("${spring.security.oauth2.client.registration.google.scope[0]}")
    private String calendarScope;

    @Value("${spring.security.oauth2.client.registration.google.client-secret}")
    private String privateKey;

    @Value("${spring.security.oauth2.client.registration.google.client-id}")
    private String serviceAccountEmail;

    @Value("${spring.GoogleJwt.tokenExpiration}")
    private long tokenExpiration;

    @Value("${spring.security.oauth2.client.provider.google.token-uri}")
    private String tokenUri;

    public void processGoogleToken() {

        var jwt = generateGoogleJwtToken();

        var params = "grant_type=" + encode("urn:ietf:params:oauth:grant-type:jwt-bearer", StandardCharsets.UTF_8)
                + "&assertion=" + encode(jwt, StandardCharsets.UTF_8);

        var googleAccessToken = restClient.post()
                .uri(tokenUri)
                .contentType(MediaType.APPLICATION_FORM_URLENCODED)
                .body(params)
                .retrieve()
                .body(new ParameterizedTypeReference<>() {});

    }

//    Decode the private key from the initial PEM format & encode into base64
    public PrivateKey loadPrivateKeyFromPem(String pem) throws InvalidKeySpecException, NoSuchAlgorithmException {

//        clean up into a clean base-64 key string
        var privateKeyPem = privateKey
                .replace("-----BEGIN PRIVATE KEY-----", "")
                .replace("-----END PRIVATE KEY-----", "")
                .replace("\\n", "");

//        Decodes the cleaned key string into raw binary
//        private key was base64-encoded, and this gives you the actual byte representation.
        var keyBytes = Base64.getDecoder().decode(privateKeyPem);

//        Creates a key factory for the RSA algorithm.
//        This factory can convert key specs into real PrivateKey instances.
        PKCS8EncodedKeySpec spec = new PKCS8EncodedKeySpec(keyBytes);
        KeyFactory keyFactory = KeyFactory.getInstance("RSA");
        return keyFactory.generatePrivate(spec);
    }

    @SneakyThrows //First time using @Todo check usage
    public String generateGoogleJwtToken(){

        var newPrivateKey = loadPrivateKeyFromPem(privateKey);

        return Jwts.builder()
                .claim("iss", serviceAccountEmail)
                .claim("scope", calendarScope)
                .claim("aud", tokenUri)
                .issuedAt(new Date())
                .expiration(new Date(System.currentTimeMillis() + 1000 * tokenExpiration))
                .signWith(newPrivateKey, Jwts.SIG.RS256)
                .compact();
    }
}
