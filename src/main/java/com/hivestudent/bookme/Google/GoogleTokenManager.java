package com.hivestudent.bookme.Google;

import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

@Component
@RequiredArgsConstructor
public class GoogleTokenManager {

    private final GoogleAuthService googleAuthService;

    private GoogleAccessToken cachedToken;

    public String getAccessToken() {
        if (cachedToken == null || cachedToken.isExpired()) {
            cachedToken = googleAuthService.processGoogleToken();
        }
        return cachedToken.getToken();
    }
}
