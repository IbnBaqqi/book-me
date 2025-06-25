package com.hivestudent.bookme.OAuth;

import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.net.URI;

@RequiredArgsConstructor
@RestController
@RequestMapping
public class OAuthController {

    private final OAuthService oAuthService;

    @Value("${spring.security.oauth2.client.registration.42-intra.client-id}")
    private String clientId;

    @Value("${spring.security.oauth2.client.registration.42-intra.redirect-uri}")
    private String redirectUri;

    //expose a route
    @GetMapping("/oauth/login")
    public ResponseEntity<Void> redirectTo42() {
        String url = "https://api.intra.42.fr/oauth/authorize"
                + "?client_id=" + clientId
                + "&redirect_uri=" + redirectUri
                + "&response_type=code"
                + "&scope=public";
                //+ "&state=" + generateRandomState();

        return ResponseEntity.status(HttpStatus.FOUND).location(URI.create(url)).build();
    }

    @GetMapping("/callback")
    public ResponseEntity<String> callback(@RequestParam String code, @RequestParam(required = false) String state) {
        var token = oAuthService.processOAuthCallback(code);
        return ResponseEntity.ok(token);
    }
}
