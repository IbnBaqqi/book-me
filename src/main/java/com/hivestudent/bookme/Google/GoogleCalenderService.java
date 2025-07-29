package com.hivestudent.bookme.Google;

import com.hivestudent.bookme.entities.Reservation;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.ParameterizedTypeReference;
import org.springframework.http.MediaType;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestClient;

import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;
import java.util.concurrent.CompletableFuture;

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
    public CompletableFuture<String> createGoogleEvent(Reservation reservation) {
        Logger log = LoggerFactory.getLogger(GoogleCalenderService.class);
        try {
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

            var response = restClient.post()
                    .uri(calendarUri + "/calendars/{calendarId}/events", calendarId)
                    .header("Authorization", "Bearer " + token)
                    .contentType(MediaType.APPLICATION_JSON)
                    .body(event)
                    .retrieve()
                    .body(new ParameterizedTypeReference<Event>() {});

            if (response == null || response.getId() == null) {
                throw new IllegalStateException("Google Calendar event creation failed");
            }

            return CompletableFuture.completedFuture(response.getId());

        } catch (Exception e) {
            log.error("Failed to create calendar event", e);
            return CompletableFuture.completedFuture(null);
//            throw new RuntimeException(e);
        }
    }

    @Async
    public void deleteGoogleEvent(String eventId) {

        var token = googleTokenManager.getAccessToken();

        restClient.delete()
                .uri(calendarUri + "/calendars/{calendarId}/events/{eventId}", calendarId, eventId)
                .header("Authorization", "Bearer " + token)
                .retrieve()
                .toBodilessEntity();
    }
}
