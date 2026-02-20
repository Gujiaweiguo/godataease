package io.dataease.embedded.util;

import com.auth0.jwt.JWT;
import com.auth0.jwt.JWTCreator;
import com.auth0.jwt.algorithms.Algorithm;
import io.dataease.auth.bo.TokenUserBO;
import io.dataease.utils.IDUtils;
import org.apache.commons.lang3.RandomStringUtils;

public class EmbeddedTokenUtil {

    private static final long TOKEN_EXPIRE_TIME = 86400000L;

    public static String generateAppId() {
        return "app_" + IDUtils.snowID();
    }

    public static String generateAppSecret(Integer length) {
        if (length == null || length <= 0) {
            length = 16;
        }
        return RandomStringUtils.randomAlphanumeric(length);
    }

    public static String generateToken(String appId, String appSecret, Long userId, Long orgId) {
        Algorithm algorithm = Algorithm.HMAC256(appSecret);
        JWTCreator.Builder builder = JWT.create();
        builder.withClaim("uid", userId);
        builder.withClaim("oid", orgId);
        builder.withClaim("appId", appId);
        builder.withClaim("exp", System.currentTimeMillis() + TOKEN_EXPIRE_TIME);
        return builder.sign(algorithm);
    }

    public static TokenUserBO validateEmbeddedToken(String token, String appSecret) {
        Algorithm algorithm = Algorithm.HMAC256(appSecret);
        JWT.require(algorithm).build().verify(token);
        Long userId = JWT.decode(token).getClaim("uid").asLong();
        Long orgId = JWT.decode(token).getClaim("oid").asLong();
        return new TokenUserBO(userId, orgId);
    }

    public static boolean isValidToken(String token, String appSecret) {
        try {
            validateEmbeddedToken(token, appSecret);
            return true;
        } catch (Exception e) {
            return false;
        }
    }
}
