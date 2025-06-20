package com.hivestudent.bookme.OAuth;

import lombok.AllArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@AllArgsConstructor
@RestController
@RequestMapping
public class OAuthController {

    private final OAuthService oAuthService;

    @GetMapping("/callback")
    public ResponseEntity<String> callback(@RequestParam String code, @RequestParam(required = false) String state) {
        oAuthService.processOAuthCallback(code);
        return ResponseEntity.ok("Login successful. You can close this tab.");
    }
}
