package com.hivestudent.bookme.Google;

import com.hivestudent.bookme.entities.Reservation;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.MediaType;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestClient;

import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;

@Service
@RequiredArgsConstructor
public class GoogleCalenderService {

    private final RestClient restClient;
    private final GoogleTokenManager googleTokenManager;

    @Value("${spring.security.oauth2.client.provider.google.calendar-uri}")
    private String calendarUri;

    @Value("${spring.security.oauth2.client.provider.google.calendar-id}")
    private String calendarId;

    @Async
    public void createGoogleEvent(Reservation reservation) {

        var token = googleTokenManager.getAccessToken();

        DateTimeFormatter formatter = DateTimeFormatter.ISO_OFFSET_DATE_TIME;

        String start = reservation.getStartTime()
                .atOffset(ZoneOffset.ofHours(3))
                .format(formatter);

        String end = reservation.getEndTime()
                .atOffset(ZoneOffset.ofHours(3))
                .format(formatter);

        var event = GoogleEventRequest.builder()
                .summary(reservation.getRoom().getName() + " Room")
                .description("Created via BookMe")
                .start(new GoogleEventRequest.DateTimeObject(start))
                .end(new GoogleEventRequest.DateTimeObject(end))
                .build();

        restClient.post()
                .uri(calendarUri + "/calendars/{calendarId}/events", calendarId)
                .header("Authorization", "Bearer " + token)
                .contentType(MediaType.APPLICATION_JSON)
                .body(event)
                .retrieve()
                .toBodilessEntity();
    }
}
