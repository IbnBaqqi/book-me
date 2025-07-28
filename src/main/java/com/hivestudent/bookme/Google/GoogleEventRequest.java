package com.hivestudent.bookme.Google;

import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class GoogleEventRequest {

    private String summary;
    private String description;
    private DateTimeObject start;
    private DateTimeObject end;

    @Data
    @Builder
    public static class DateTimeObject {
        private String dateTime;
    }
}
