package com.hivestudent.bookme.Google;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Getter;
import lombok.Setter;

import java.time.Instant;

@Setter
@Getter
public class GoogleAccessToken {

    @JsonProperty("access_token")
    private String token;

    private Instant createdAt;

    @JsonProperty("expires_in")
    private int expiresIn;

    public boolean isExpired() {
        return Instant.now().isAfter(createdAt.plusSeconds(expiresIn - 60));
    }
}
