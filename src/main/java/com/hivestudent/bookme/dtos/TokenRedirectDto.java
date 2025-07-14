package com.hivestudent.bookme.dtos;

import lombok.AllArgsConstructor;
import lombok.Data;

/**
 * Contains data for token redirect
 */
@Data
@AllArgsConstructor
public class TokenRedirectDto {
    private String token;
    private String intra;
    private String role;
}
