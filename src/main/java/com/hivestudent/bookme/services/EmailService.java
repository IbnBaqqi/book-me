package com.hivestudent.bookme.services;

import jakarta.mail.MessagingException;

import java.io.IOException;

public interface EmailService {
    void sendEmail(String email, String room, String date) throws MessagingException, IOException;
}
