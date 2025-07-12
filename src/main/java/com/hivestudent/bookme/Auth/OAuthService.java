package com.hivestudent.bookme.Auth;

import com.hivestudent.bookme.dao.UserRepository;
import com.hivestudent.bookme.dtos.FortyTwoTokenResponse;
import com.hivestudent.bookme.dtos.IntraUserDto;
import com.hivestudent.bookme.entities.Role;
import com.hivestudent.bookme.entities.User;
import com.hivestudent.bookme.exceptions.RestTemplateErrorHandler;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.*;
import org.springframework.security.access.AccessDeniedException;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.stereotype.Service;
import org.springframework.util.LinkedMultiValueMap;
import org.springframework.web.client.HttpClientErrorException;
import org.springframework.web.client.RestClientException;
import org.springframework.web.client.RestTemplate;

import java.util.List;
import java.util.Objects;


@Service
@RequiredArgsConstructor
public class OAuthService {

    private final UserRepository userRepository;
    private final JwtService jwtService;
    private final RestTemplate restTemplate; //manual bean in appConfig

    @Value("${spring.security.oauth2.client.registration.42-intra.client-id}")
    private String clientId;

    @Value("${spring.security.oauth2.client.registration.42-intra.client-secret}")
    private String clientSecret;

    @Value("${spring.security.oauth2.client.provider.42-intra.user-info-uri}")
    private String userInfoUrl;

    @Value("${spring.security.oauth2.client.registration.42-intra.redirect-uri}")
    private String redirectUri;

    @Value("${spring.security.oauth2.client.provider.42-intra.token-uri}")
    private String tokenUrl;

    public String processOAuthCallback(String code) {
//        Step 1: Exchange code for accessToken
        restTemplate.setErrorHandler(new RestTemplateErrorHandler());
//        create request body parameters
        var params = new LinkedMultiValueMap<String, String>();
        params.add("grant_type", "authorization_code");
        params.add("client_id", clientId);
        params.add("client_secret", clientSecret);
        params.add("code", code);
        params.add("redirect_uri", redirectUri);

//        set content-type in headers
        var headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_FORM_URLENCODED);

//        Wrap into HttpEntity (a request Object)
        var request = new HttpEntity<>(params, headers);

//        Send an http post request and parse the json into a ResponseDto
        ResponseEntity<FortyTwoTokenResponse> tokenResponse;
        try {
            tokenResponse = restTemplate.postForEntity(tokenUrl, request, FortyTwoTokenResponse.class);
        } catch (HttpClientErrorException e) {
            throw new RestClientException("Error " + e.getStatusCode() + " : " + e.getMessage());
        }

//        Get accessToken from the Response
        var token = tokenResponse.getBody();
        assert token != null;
        var accessToken = token.getAccessToken();

//        Stage 2 Fetch Current User from
        var user = getCurrentUser(accessToken);

//        Generate jwt
        return jwtService.generateToken(user);
    }

//    move to userService later
    public User getCurrentUser(String accessToken) {
        var userHeaders = new HttpHeaders();
        userHeaders.setBearerAuth(accessToken);

        var userRequest = new HttpEntity<>(userHeaders);

        // Extract User Data
        ResponseEntity<IntraUserDto> userResponse;
        try {
            userResponse = restTemplate.exchange(userInfoUrl, HttpMethod.GET, userRequest, IntraUserDto.class);
        } catch (HttpClientErrorException e) {
            throw new RestClientException("Error " + e.getStatusCode() + " : " + e.getMessage());
        }

        var userData = Objects.requireNonNull(userResponse.getBody(), "Failed to get user info");

        // Step 3: Find or create user
        String email = userData.getEmail();
        String name = userData.getName();
        List<IntraUserDto.CampusUsers> campus = userData.getCampus();

        // Check if account primarily belongs to Hive campus
        var hive = campus.stream()
                .filter(camp -> camp.getId() == 13 && camp.isPrimary())
                .findFirst()
                .orElse(null);
        if (hive == null)
            throw new AccessDeniedException("Only Helsinki Campus Student Allowed");

//        check if user is already in db, if not create new user, assign role & slap it on db
        return userRepository.findByEmail(email)
                .orElseGet(() -> {
                    User newUser = new User();
                    newUser.setEmail(email);
                    newUser.setName(name);
                    newUser.setRole(userData.isStaff() ? Role.STAFF : Role.STUDENT);
                    return userRepository.save(newUser);
                });
    }

    //overloaded method, get user from App Context
    public User getCurrentUser() {
        var authentication = SecurityContextHolder.getContext().getAuthentication();
        var email = (String) authentication.getPrincipal();

        return userRepository.findByEmail(email).orElse(null);
    }
}
