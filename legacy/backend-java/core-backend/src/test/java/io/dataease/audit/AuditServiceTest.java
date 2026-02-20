package io.dataease.audit;

import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import io.dataease.audit.dao.auto.mapper.AuditLogMapper;
import io.dataease.audit.dao.auto.mapper.LoginFailureMapper;
import io.dataease.audit.entity.AuditLog;
import io.dataease.audit.entity.LoginFailure;
import io.dataease.audit.service.impl.AuditServiceImpl;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.time.LocalDateTime;
import java.util.Arrays;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyLong;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
public class AuditServiceTest {

    @Mock
    private AuditLogMapper auditLogMapper;

    @Mock
    private LoginFailureMapper loginFailureMapper;

    @InjectMocks
    private AuditServiceImpl auditService;

    @Test
    public void createAuditLogSetsDefaults() {
        AuditLog log = new AuditLog();
        log.setActionType("USER_ACTION");
        log.setActionName("LOGIN");
        log.setOperation("LOGIN");

        AuditLog result = auditService.createAuditLog(log);

        verify(auditLogMapper, times(1)).insert(log);
        assertNotNull(result.getCreateTime());
        assertEquals("SUCCESS", result.getStatus());
        assertEquals("127.0.0.1", result.getIpAddress());
    }

    @Test
    public void queryAuditLogsReturnsPage() {
        Page<AuditLog> page = new Page<>(1, 10);
        AuditLog log = new AuditLog();
        log.setId(1L);
        page.setRecords(Arrays.asList(log));
        page.setTotal(1);

        when(auditLogMapper.selectPage(any(Page.class), any())).thenReturn(page);

        IPage<AuditLog> result = auditService.queryAuditLogs(
                null, null, null, null, null, null, null, 1, 10
        );

        assertEquals(1, result.getTotal());
        assertEquals(1, result.getRecords().size());
    }

    @Test
    public void recordLoginFailureInserts() {
        auditService.recordLoginFailure("user", "127.0.0.1", "agent", "bad password");
        verify(loginFailureMapper, times(1)).insert(any(LoginFailure.class));
    }

    @Test
    public void deleteAuditLogsBeforeDateRemovesEach() {
        AuditLog log1 = new AuditLog();
        log1.setId(1L);
        AuditLog log2 = new AuditLog();
        log2.setId(2L);

        when(auditLogMapper.selectList(any())).thenReturn(Arrays.asList(log1, log2));

        auditService.deleteAuditLogsBeforeDate(LocalDateTime.now().minusDays(1));

        verify(auditLogMapper, times(1)).deleteById(1L);
        verify(auditLogMapper, times(1)).deleteById(2L);
    }
}
