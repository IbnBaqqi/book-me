package com.hivestudent.bookme.dtos;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.util.List;

@Data
public class IntraUserDto {

    private String email;

    @JsonProperty("login")
    private String name;

    @JsonProperty("staff?")
    private boolean staff;

    @JsonProperty("campus")
    private List<Campus> campus;

    @Data
    public static class Campus {
        private int id;
        private String name;
    }
}
