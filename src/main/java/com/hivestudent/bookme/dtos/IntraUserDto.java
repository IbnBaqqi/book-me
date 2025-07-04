package com.hivestudent.bookme.dtos;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

@Data
public class IntraUserDto {

    private String email;

    @JsonProperty("login")
    private String name;

    @JsonProperty("staff?")
    private boolean staff;
}
