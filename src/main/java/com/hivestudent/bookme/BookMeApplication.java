package com.hivestudent.bookme;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableAsync;

@SpringBootApplication
@EnableAsync
public class BookMeApplication {

    public static void main(String[] args) {
        SpringApplication.run(BookMeApplication.class, args);
    }

}
