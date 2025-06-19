package com.hivestudent.bookme.dao;

import com.hivestudent.bookme.entities.Room;
import org.springframework.data.jpa.repository.JpaRepository;

public interface RoomRepository extends JpaRepository<Room, Long> {
}