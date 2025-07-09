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

    @JsonProperty("campus_users")
    private List<CampusUsers> campus;

    @Data
    public static class CampusUsers {

        @JsonProperty("campus_id")
        private int id;

        @JsonProperty("is_primary")
        private boolean primary;
    }
}
