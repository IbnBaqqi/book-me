package com.hivestudent.bookme.exceptions;

import org.springframework.http.HttpMethod;
import org.springframework.http.client.ClientHttpResponse;
import org.springframework.web.client.HttpClientErrorException;
import org.springframework.web.client.HttpServerErrorException;
import org.springframework.web.client.ResponseErrorHandler;

import java.io.IOException;
import java.net.URI;

import static org.springframework.http.HttpStatus.*;

public class RestTemplateErrorHandler implements ResponseErrorHandler {

    @Override
    public boolean hasError(ClientHttpResponse response) throws IOException {

        var statusCode = response.getStatusCode();
        return statusCode.is4xxClientError() || statusCode.is5xxServerError();
    }


    @Override
    public void handleError(URI url, HttpMethod method, ClientHttpResponse response) throws IOException {
        var status = response.getStatusCode();
        var text = response.getStatusText();

        if (status.is4xxClientError()) {

            throw new HttpClientErrorException(status, text);
        }
        else if (status.is5xxServerError()) {

            if (status == SERVICE_UNAVAILABLE)
                throw new HttpClientErrorException(SERVICE_UNAVAILABLE, "42 Intra service is currently unavailable.");
            else
                throw new HttpServerErrorException(status, text);
        }
    }
}
