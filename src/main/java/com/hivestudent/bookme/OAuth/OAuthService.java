package com.hivestudent.bookme.OAuth;

import com.hivestudent.bookme.dao.UserRepository;
import com.hivestudent.bookme.entities.Role;
import com.hivestudent.bookme.entities.User;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpMethod;
import org.springframework.http.MediaType;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.stereotype.Service;
import org.springframework.util.LinkedMultiValueMap;
import org.springframework.web.client.RestTemplate;

import java.util.Map;

@Service
@RequiredArgsConstructor
public class OAuthService {

    private final UserRepository userRepository;
    private final JwtService jwtService;
    RestTemplate restTemplate = new RestTemplate();

    @Value("${spring.security.oauth2.client.registration.42-intra.client-id}")
    private String clientId;

    @Value("${spring.security.oauth2.client.registration.42-intra.client-secret}")
    private String clientSecret;

    @Value("${spring.security.oauth2.client.registration.42-intra.redirect-uri}")
    private String redirectUri;

    @Value("${spring.security.oauth2.client.provider.42-intra.token-uri}")
    private String tokenUrl;

    public String processOAuthCallback(String code) {
//        Step 1: Exchange code for accessToken

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

//        Send a http post request and parse the json into a generic Map
        // @Todo change Map.class to tokenResponseDto later
        var tokenResponse = restTemplate.postForEntity(tokenUrl, request, Map.class);

//        Get accessToken from the Map
        String accessToken = (String) tokenResponse.getBody().get("access_token");
        System.out.println(accessToken);

//        Stage 2 Fetch Current User from
        var user = getCurrentUser(accessToken);

//
        return jwtService.generateToken(user);
    }

//    move to userService later
    public User getCurrentUser(String accessToken) {
        var userInfoUrl = "https://api.intra.42.fr/v2/me";
        var userHeaders = new HttpHeaders();
        userHeaders.setBearerAuth(accessToken);

        var userRequest = new HttpEntity<>(userHeaders);

        // @Todo change Map.class to IntraUserDto later
        var userResponse = restTemplate.exchange(userInfoUrl, HttpMethod.GET, userRequest, Map.class);
        Map<String, Object> userData = userResponse.getBody();

        // Step 3: Find or create user
        String email = (String) userData.get("email");
        String name = (String) userData.get("displayname");

//        @Todo check if user is staff and make role to be staff
        // default role
        return userRepository.findByEmail(email)
                .orElseGet(() -> {
                    User newUser = new User();
                    newUser.setEmail(email);
                    newUser.setName(name);
                    newUser.setRole(Role.STUDENT); // default role
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
