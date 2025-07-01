package com.hivestudent.bookme.dtos;

import lombok.AllArgsConstructor;
import lombok.Data;

import java.time.LocalDateTime;
import java.util.List;

@Data
@AllArgsConstructor
public class ReservedDto {
    private long roomId;
    private String roomName;
    private List<Slot> slots;

    @Data
    @AllArgsConstructor
    public static class Slot {
        private long id;
        private LocalDateTime start;
        private LocalDateTime end;
        private String bookedBy; // null if not staff
    }
}
