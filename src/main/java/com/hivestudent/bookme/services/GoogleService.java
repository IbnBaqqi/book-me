package com.hivestudent.bookme.services;

import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.security.Keys;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.util.Date;

@Service
@RequiredArgsConstructor
public class GoogleService {

    @Value("${spring.security.oauth2.client.registration.google.scope[0]}")
    private String calendarScope;

    @Value("${spring.security.oauth2.client.registration.google.client-secret}")
    private String secretKey;

    @Value("${spring.security.oauth2.client.registration.google.client-id}")
    private String serviceAccountEmail;

    @Value("${spring.GoogleJwt.tokenExpiration}")
    private long tokenExpiration;

    @Value("${spring.security.oauth2.client.provider.google.token-uri}")
    private String tokenUri;

    public String processGoogleToken() {

        var jwt = generateGoogleJwtToken();
        return null;
    }

    public String generateGoogleJwtToken() {
        return Jwts.builder()
                .claim("iss", serviceAccountEmail)
                .claim("scope", calendarScope)
                .claim("aud", tokenUri)
                .issuedAt(new Date())
                .expiration(new Date(System.currentTimeMillis() + 1000 * tokenExpiration))
                .signWith(Keys.hmacShaKeyFor(secretKey.getBytes()))
                .compact();
    }
}
