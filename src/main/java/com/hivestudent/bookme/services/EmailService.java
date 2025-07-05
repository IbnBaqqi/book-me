package com.hivestudent.bookme.services;

import jakarta.mail.MessagingException;

public interface EmailService {
    void sendEmail(String email) throws MessagingException;
}
