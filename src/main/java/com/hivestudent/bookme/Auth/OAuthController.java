package com.hivestudent.bookme.Auth;

import jakarta.servlet.http.HttpServletResponse;
import jakarta.servlet.http.HttpSession;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.io.IOException;
import java.net.URI;
import java.util.UUID;

@RequiredArgsConstructor
@RestController
@RequestMapping("/oauth")
public class OAuthController {

    private final OAuthService oAuthService;

    @Value("${spring.security.oauth2.client.registration.42-intra.client-id}")
    private String clientId;

    @Value("${spring.security.oauth2.client.registration.42-intra.redirect-uri}")
    private String redirectUri;

    @Value("${spring.redirect.token_url}")
    private String tokenRedirect;

    //expose a route
    @GetMapping("/login")
    public ResponseEntity<Void> redirectTo42(HttpSession session) {

        // state to prevent csrf
        String state = UUID.randomUUID().toString();
        session.setAttribute("oauth_state", state);

        String url = "https://api.intra.42.fr/oauth/authorize"
                + "?client_id=" + clientId
                + "&redirect_uri=" + redirectUri
                + "&response_type=code"
                + "&scope=public"
                + "&state=" + state;

        return ResponseEntity.status(HttpStatus.FOUND).location(URI.create(url)).build();
    }

    @GetMapping("/callback")
    public void callback(@RequestParam String code, HttpServletResponse response, @RequestParam(required = false) String state, HttpSession session) throws IOException {

        String savedState = (String) session.getAttribute("oauth_state");

        if (savedState == null || !savedState.equals(state)) {
            response.sendError(HttpServletResponse.SC_FORBIDDEN, "Invalid or missing state — possible CSRF attack.");
            return;
        }

        session.removeAttribute("oauth_state");

        var token = oAuthService.processOAuthCallback(code);
        response.sendRedirect(tokenRedirect + token);
    }
}
