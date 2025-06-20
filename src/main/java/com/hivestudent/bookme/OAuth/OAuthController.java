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
        var token = oAuthService.processOAuthCallback(code);
        return ResponseEntity.ok(token);
    }
}
