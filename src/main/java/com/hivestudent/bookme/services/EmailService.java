package com.hivestudent.bookme.services;

import jakarta.mail.MessagingException;

import java.io.IOException;

public interface EmailService {
    void sendConfirmation(String email, String room, String date) throws MessagingException, IOException;
}
