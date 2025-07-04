package com.hivestudent.bookme.common;


import io.github.bucket4j.Bucket;
import jakarta.servlet.*;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.AllArgsConstructor;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.time.Duration;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Component
@AllArgsConstructor
public class RateLimitFilter implements Filter {

    private final Map<String, Bucket> buckets = new ConcurrentHashMap<>();

    @Override
    public void doFilter(ServletRequest request, ServletResponse response, FilterChain filterChain) throws IOException, ServletException {

        HttpServletRequest httpRequest = (HttpServletRequest) request;
        String ipAddress = httpRequest.getRemoteAddr();
        Bucket bucket = buckets.computeIfAbsent(ipAddress, this::newBucket);
        if (bucket.tryConsume(1)) {
            filterChain.doFilter(request, response);
        } else {
            HttpServletResponse httpResponse = (HttpServletResponse) response;
            httpResponse.setStatus(429);
            httpResponse.getWriter().write("Too many requests");
        }
    }

    private Bucket newBucket(String key) {
        return Bucket.builder()
                .addLimit(limit -> limit
                        .capacity(1)
                        .refillGreedy(1, Duration.ofSeconds(1)))
                .build();
    }
}
