package io.dataease.embedded.server;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.auth0.jwt.JWT;
import io.dataease.api.permissions.embedded.dto.EmbeddedCreator;
import io.dataease.api.permissions.embedded.dto.EmbeddedEditor;
import io.dataease.api.permissions.embedded.dto.EmbeddedOrigin;
import io.dataease.api.permissions.embedded.dto.EmbeddedResetRequest;
import io.dataease.api.permissions.embedded.vo.EmbeddedGridVO;
import io.dataease.audit.annotation.AuditLog;
import io.dataease.audit.constant.AuditConstants;
import io.dataease.auth.bo.TokenUserBO;
import io.dataease.embedded.dao.auto.entity.CoreEmbedded;
import io.dataease.embedded.dao.auto.mapper.CoreEmbeddedMapper;
import io.dataease.embedded.util.EmbeddedTokenUtil;
import io.dataease.exception.DEException;
import io.dataease.model.KeywordRequest;
import io.dataease.utils.AuthUtils;
import jakarta.annotation.Resource;
import org.apache.commons.collections4.CollectionUtils;
import org.apache.commons.lang3.StringUtils;
import org.springframework.web.bind.annotation.*;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/embedded")
public class EmbeddedServer {

    @Resource
    private CoreEmbeddedMapper embeddedMapper;

    @PostMapping("/pager/{goPage}/{pageSize}")
    public IPage<EmbeddedGridVO> queryGrid(@PathVariable("goPage") int goPage, @PathVariable("pageSize") int pageSize, @RequestBody KeywordRequest request) {
        QueryWrapper<CoreEmbedded> queryWrapper = new QueryWrapper<>();
        if (StringUtils.isNotBlank(request.getKeyword())) {
            queryWrapper.like("name", request.getKeyword());
        }
        queryWrapper.orderByDesc("create_time");
        Page<CoreEmbedded> page = new Page<>(goPage, pageSize);
        IPage<CoreEmbedded> resultPage = embeddedMapper.selectPage(page, queryWrapper);
        IPage<EmbeddedGridVO> voPage = new Page<>(goPage, pageSize);
        List<EmbeddedGridVO> voList = new ArrayList<>();
        if (CollectionUtils.isNotEmpty(resultPage.getRecords())) {
            for (CoreEmbedded embedded : resultPage.getRecords()) {
                EmbeddedGridVO vo = new EmbeddedGridVO();
                vo.setId(embedded.getId());
                vo.setName(embedded.getName());
                vo.setAppId(embedded.getAppId());
                vo.setAppSecret(maskAppSecret(embedded.getAppSecret()));
                vo.setDomain(embedded.getDomain());
                vo.setSecretLength(embedded.getSecretLength());
                voList.add(vo);
            }
        }
        voPage.setRecords(voList);
        voPage.setTotal(resultPage.getTotal());
        return voPage;
    }

    @PostMapping("/create")
    @AuditLog(
        actionType = AuditConstants.ACTION_TYPE_USER_ACTION,
        actionName = "CREATE_EMBEDDED_APP",
        resourceType = "EMBEDDED_APP"
    )
    public void create(@RequestBody EmbeddedCreator creator) {
        String appId = EmbeddedTokenUtil.generateAppId();
        String appSecret = EmbeddedTokenUtil.generateAppSecret(creator.getSecretLength());
        CoreEmbedded embedded = new CoreEmbedded();
        embedded.setName(creator.getName());
        embedded.setAppId(appId);
        embedded.setAppSecret(appSecret);
        embedded.setDomain(creator.getDomain());
        embedded.setSecretLength(creator.getSecretLength() != null ? creator.getSecretLength() : 16);
        embedded.setCreateTime(System.currentTimeMillis());
        embedded.setUpdateBy(AuthUtils.getUser().getUserId().toString());
        embedded.setUpdateTime(System.currentTimeMillis());
        embeddedMapper.insert(embedded);
    }

    @PostMapping("/edit")
    @AuditLog(
        actionType = AuditConstants.ACTION_TYPE_USER_ACTION,
        actionName = "UPDATE_EMBEDDED_APP",
        resourceType = "EMBEDDED_APP"
    )
    public void edit(@RequestBody EmbeddedEditor editor) {
        if (editor.getId() == null) {
            DEException.throwException("ID不能为空");
        }
        CoreEmbedded embedded = embeddedMapper.selectById(editor.getId());
        if (embedded == null) {
            DEException.throwException("嵌入式应用不存在");
        }
        embedded.setName(editor.getName());
        if (StringUtils.isNotBlank(editor.getDomain())) {
            embedded.setDomain(editor.getDomain());
        }
        if (editor.getSecretLength() != null) {
            embedded.setSecretLength(editor.getSecretLength());
        }
        embedded.setUpdateBy(AuthUtils.getUser().getUserId().toString());
        embedded.setUpdateTime(System.currentTimeMillis());
        embeddedMapper.updateById(embedded);
    }

    @PostMapping("/delete/{id}")
    @AuditLog(
        actionType = AuditConstants.ACTION_TYPE_USER_ACTION,
        actionName = "DELETE_EMBEDDED_APP",
        resourceType = "EMBEDDED_APP"
    )
    public void delete(@PathVariable("id") Long id) {
        if (id == null) {
            DEException.throwException("ID不能为空");
        }
        embeddedMapper.deleteById(id);
    }

    @PostMapping("/batchDelete")
    @AuditLog(
        actionType = AuditConstants.ACTION_TYPE_USER_ACTION,
        actionName = "BATCH_DELETE_EMBEDDED_APP",
        resourceType = "EMBEDDED_APP"
    )
    public void batchDelete(@RequestBody List<Long> ids) {
        if (CollectionUtils.isEmpty(ids)) {
            DEException.throwException("ID列表不能为空");
        }
        embeddedMapper.deleteBatchIds(ids);
    }

    @PostMapping("/reset")
    @AuditLog(
        actionType = AuditConstants.ACTION_TYPE_USER_ACTION,
        actionName = "RESET_EMBEDDED_APP_SECRET",
        resourceType = "EMBEDDED_APP"
    )
    public void reset(@RequestBody EmbeddedResetRequest request) {
        if (request.getId() == null) {
            DEException.throwException("ID不能为空");
        }
        CoreEmbedded embedded = embeddedMapper.selectById(request.getId());
        if (embedded == null) {
            DEException.throwException("嵌入式应用不存在");
        }
        String newSecret = StringUtils.isNotBlank(request.getAppSecret())
            ? request.getAppSecret()
            : EmbeddedTokenUtil.generateAppSecret(embedded.getSecretLength());
        embedded.setAppSecret(newSecret);
        embedded.setUpdateBy(AuthUtils.getUser().getUserId().toString());
        embedded.setUpdateTime(System.currentTimeMillis());
        embeddedMapper.updateById(embedded);
    }

    @GetMapping("/domainList")
    public List<String> domainList() {
        QueryWrapper<CoreEmbedded> queryWrapper = new QueryWrapper<>();
        queryWrapper.select("DISTINCT domain");
        queryWrapper.isNotNull("domain");
        queryWrapper.ne("domain", "");
        List<CoreEmbedded> list = embeddedMapper.selectList(queryWrapper);
        List<String> domains = new ArrayList<>();
        if (CollectionUtils.isNotEmpty(list)) {
            for (CoreEmbedded embedded : list) {
                if (StringUtils.isNotBlank(embedded.getDomain())) {
                    domains.add(embedded.getDomain());
                }
            }
        }
        return domains;
    }

    @PostMapping("/initIframe")
    @AuditLog(
        actionType = AuditConstants.ACTION_TYPE_DATA_ACCESS,
        actionName = "EMBEDDED_IFRAME_ACCESS",
        resourceType = "EMBEDDED_APP"
    )
    public List<String> initIframe(@RequestBody EmbeddedOrigin origin) {
        String embeddedToken = origin.getToken();
        if (StringUtils.isBlank(embeddedToken)) {
            DEException.throwException("嵌入式Token不能为空");
        }
        String appId = JWT.decode(embeddedToken).getClaim("appId").asString();
        if (StringUtils.isBlank(appId)) {
            DEException.throwException("嵌入式Token无效");
        }
        QueryWrapper<CoreEmbedded> queryWrapper = new QueryWrapper<>();
        queryWrapper.eq("app_id", appId);
        CoreEmbedded embedded = embeddedMapper.selectOne(queryWrapper);
        if (embedded == null) {
            DEException.throwException("嵌入式应用不存在");
        }
        if (!EmbeddedTokenUtil.isValidToken(embeddedToken, embedded.getAppSecret())) {
            DEException.throwException("嵌入式Token无效");
        }
        String originUrl = origin.getOrigin();
        if (!isOriginAllowed(originUrl, embedded.getDomain())) {
            DEException.throwException("嵌入式来源不合法");
        }
        return parseDomains(embedded.getDomain());
    }

    @GetMapping("/getTokenArgs")
    public Map<String, Object> getTokenArgs() {
        TokenUserBO user = AuthUtils.getUser();
        Map<String, Object> result = new HashMap<>();
        result.put("userId", user.getUserId());
        result.put("orgId", user.getDefaultOid());
        return result;
    }

    @GetMapping("/limitCount")
    public int getLimitCount() {
        return 5;
    }

    private String maskAppSecret(String appSecret) {
        if (StringUtils.isBlank(appSecret)) {
            return "";
        }
        if (appSecret.length() <= 8) {
            return "********";
        }
        return appSecret.substring(0, 4) + "****" + appSecret.substring(appSecret.length() - 4);
    }

    private boolean isOriginAllowed(String origin, String domainList) {
        if (StringUtils.isBlank(domainList)) {
            return false;
        }
        String normalizedOrigin = normalizeOrigin(origin);
        if (StringUtils.isBlank(normalizedOrigin)) {
            return false;
        }
        String originHost = extractHost(normalizedOrigin);
        return parseDomains(domainList).stream().anyMatch(allowed -> {
            if (allowed.equalsIgnoreCase(normalizedOrigin)) {
                return true;
            }
            return StringUtils.isNotBlank(originHost) && allowed.equalsIgnoreCase(originHost);
        });
    }

    private List<String> parseDomains(String domainList) {
        return Arrays.stream(domainList.split("[,;\\s]+"))
            .map(this::normalizeOrigin)
            .filter(StringUtils::isNotBlank)
            .collect(Collectors.toList());
    }

    private String normalizeOrigin(String origin) {
        if (StringUtils.isBlank(origin)) {
            return "";
        }
        String normalized = origin.trim();
        while (normalized.endsWith("/")) {
            normalized = StringUtils.removeEnd(normalized, "/");
        }
        return normalized;
    }

    private String extractHost(String origin) {
        try {
            return java.net.URI.create(origin).getHost();
        } catch (IllegalArgumentException error) {
            return "";
        }
    }
}
