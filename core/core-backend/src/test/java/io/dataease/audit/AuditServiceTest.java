package io.dataease.audit;

import io.dataease.audit.entity.AuditLog;
import io.dataease.audit.service.IAuditService;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentMatchers;
import static org.junit.jupiter.api.Assertions.*;

@ExtendWith(MockitoExtension.class)
public class AuditServiceUnitTest {

    @Mock
    private IAuditService auditService;

    @InjectMocks
    private IAuditService target;

    private AuditLog testAuditLog;

    @BeforeEach
    public void setUp() {
        MockitoAnnotations.openMocks(this);
        testAuditLog = new AuditLog();
    }

    @Test
    public void testCreateAuditLog() throws Exception {
        when(auditService.createAuditLog(any())).thenReturn(testAuditLog);
        target.createAuditLog(testAuditLog);
        verify(auditService, times(1)).createAuditLog(any());
    }

    @Test
    public void testQueryAuditLogs() throws Exception {
        LocalDateTime startTime = LocalDateTime.of(2024, 1, 1, 0, 0);
        LocalDateTime endTime = LocalDateTime.of(2024, 1, 2, 0, 0);
        
        AuditLog log1 = new AuditLog();
        log1.setId(1L);
        log1.setUsername("user1");
        log1.setActionType("USER_ACTION");
        log1.setActionName("LOGIN");
        log1.setStatus("SUCCESS");
        log1.setCreateTime(startTime);
        
        AuditLog log2 = new AuditLog();
        log2.setId(2L);
        log2.setUsername("user1");
        log2.setActionType("USER_ACTION");
        log2.setActionName("LOGOUT");
        log2.setStatus("SUCCESS");
        log2.setCreateTime(endTime);
        
        List<AuditLog> mockLogs = Arrays.asList(log1, log2);
        
        when(auditService.queryAuditLogs(
                any(), any(), any(), any(), any(), any(), 
                startTime, endTime, 1, 10
        )).thenReturn(new com.baomidou.mybatisplus.core.metadata.IPage<AuditLog>(mockLogs, 0, 2, 2));
        
        var result = auditService.queryAuditLogs(
                null, null, null, null, startTime, endTime, 1, 10
        );
        
        assertNotNull(result);
        assertEquals(2, result.getTotal());
        assertEquals(2, result.getRecords().size());
    }

    @Test
    public void testDeleteAuditLogsBeforeDate() throws Exception {
        LocalDateTime cutoffDate = LocalDateTime.of(2023, 12, 31, 23, 59, 59);
        
        when(auditService.deleteAuditLogsBeforeDate(eq(cutoffDate)))
                .then(invocation -> {}).Void;
        
        verify(auditService, times(1)).deleteAuditLogsBeforeDate(eq(cutoffDate));
    }

    @Test
    public void testExportAuditLogs() throws Exception {
        List<Long> ids = Arrays.asList(1L, 2L);
        
        when(auditService.exportAuditLogs(eq(ids), eq("csv")))
                .then(invocation -> {}).Void;
        
        verify(auditService, times(1)).exportAuditLogs(eq(ids), eq("csv"));
    }
}

    @Test
    public void testCreateAuditLog() throws Exception {
        when(auditService.createAuditLog(any())).thenReturn(testAuditLog);
        target.createAuditLog(testAuditLog);
        verify(auditService, times(1)).createAuditLog(any());
    }

    @Test
    public void testQueryAuditLogs() throws Exception {
        LocalDateTime startTime = LocalDateTime.of(2024, 1, 1, 0, 0);
        LocalDateTime endTime = LocalDateTime.of(2024, 1, 2, 0, 0);
        
        AuditLog log1 = new AuditLog();
        log1.setId(1L);
        log1.setUsername("user1");
        log1.setActionType("USER_ACTION");
        log1.setActionName("LOGIN");
        log1.setStatus("SUCCESS");
        log1.setCreateTime(startTime);
        
        AuditLog log2 = new AuditLog();
        log2.setId(2L);
        log2.setUsername("user1");
        log2.setActionType("USER_ACTION");
        log2.setActionName("LOGOUT");
        log2.setStatus("SUCCESS");
        log2.setCreateTime(endTime);
        
        List<AuditLog> mockLogs = Arrays.asList(log1, log2);
        
        when(auditService.queryAuditLogs(
                any(), any(), any(), any(), any(), any(), 
                startTime, endTime, 1, 10
        )).thenReturn(mockLogs);
        
        var result = auditService.queryAuditLogs(
                null, null, null, null, startTime, endTime, 1, 10
        );
        
        assertNotNull(result);
        assertEquals(2, result.getTotal());
        assertEquals(2, result.getRecords().size());
    }

    @Test
    public void testDeleteAuditLogsBeforeDate() throws Exception {
        LocalDateTime cutoffDate = LocalDateTime.of(2023, 12, 31, 23, 59, 59);
        
        when(auditService.deleteAuditLogsBeforeDate(eq(cutoffDate)))
                .then(invocation -> {}).Void;
        
        verify(auditService, times(1)).deleteAuditLogsBeforeDate(eq(cutoffDate));
    }

    @Test
    public void testExportAuditLogs() throws Exception {
        List<Long> ids = Arrays.asList(1L, 2L);
        
        when(auditService.exportAuditLogs(eq(ids), eq("csv")))
                .then(invocation -> {}).Void;
        
        verify(auditService, times(1)).exportAuditLogs(eq(ids), eq("csv"));
    }
}

    @Test
    public void testQueryAuditLogs() throws Exception {
        IAuditService auditService = mock(IAuditService.class);
        
        LocalDateTime startTime = LocalDateTime.of(2024, 1, 1, 0, 0);
        LocalDateTime endTime = LocalDateTime.of(2024, 1, 2, 0, 0);
        
        AuditLog log1 = new AuditLog();
        log1.setId(1L);
        log1.setUsername("user1");
        log1.setActionType("USER_ACTION");
        log1.setActionName("LOGIN");
        log1.setStatus("SUCCESS");
        log1.setCreateTime(startTime);
        
        AuditLog log2 = new AuditLog();
        log2.setId(2L);
        log2.setUsername("user1");
        log2.setActionType("USER_ACTION");
        log2.setActionName("LOGOUT");
        log2.setStatus("SUCCESS");
        log2.setCreateTime(endTime);
        
        List<AuditLog> mockLogs = Arrays.asList(log1, log2);
        
        when(auditService.queryAuditLogs(
                any(), any(), any(), any(), any(), any(), 
                startTime, endTime, 1, 10
        )).thenReturn(mockLogs);
        
        var result = auditService.queryAuditLogs(
                null, null, null, null, startTime, endTime, 1, 10
        );
        
        assertNotNull(result);
        assertEquals(2, result.getTotal());
        assertEquals(2, result.getRecords().size());
    }

    @Test
    public void testDeleteAuditLogsBeforeDate() throws Exception {
        IAuditService auditService = mock(IAuditService.class);
        LocalDateTime cutoffDate = LocalDateTime.of(2023, 12, 31, 23, 59);
        
        when(auditService.deleteAuditLogsBeforeDate(eq(cutoffDate)))
                .then(invocation -> {}).Void;
        
        verify(auditService, times(1)).deleteAuditLogsBeforeDate(eq(cutoffDate));
    }

    @Test
    public void testExportAuditLogs() throws Exception {
        IAuditService auditService = mock(IAuditService.class);
        List<Long> ids = Arrays.asList(1L, 2L, 3L);
        
        when(auditService.exportAuditLogs(eq(ids), eq("csv")))
                .then(invocation -> {}).Void;
        
        verify(auditService, times(1)).exportAuditLogs(eq(ids), eq("csv"));
    }
}

    @Test
    public void testCreateAuditLog() throws Exception {
        when(auditService.createAuditLog(any(AuditLog.class)))
                .thenReturn(testAuditLog);
        
        target.createAuditLog(testAuditLog);
        
        verify(auditService, times(1)).createAuditLog(any(AuditLog.class));
    }

    @Test
    public void testQueryAuditLogs() throws Exception {
        LocalDateTime startTime = LocalDateTime.of(2024, 1, 1, 0, 0);
        LocalDateTime endTime = LocalDateTime.of(2024, 1, 2, 0, 0);
        
        AuditLog log1 = new AuditLog();
        log1.setId(1L);
        log1.setUsername("user1");
        log1.setActionType("USER_ACTION");
        log1.setActionName("LOGIN");
        log1.setStatus("SUCCESS");
        log1.setCreateTime(startTime);
        
        AuditLog log2 = new AuditLog();
        log2.setId(2L);
        log2.setUsername("user1");
        log2.setActionType("USER_ACTION");
        log2.setActionName("LOGOUT");
        log2.setStatus("SUCCESS");
        log2.setCreateTime(endTime);
        
        List<AuditLog> mockLogs = Arrays.asList(log1, log2);
        
        when(auditService.queryAuditLogs(
                any(), any(), any(), any(), 
                any(), any(), startTime, endTime, 1, 10
        )).thenReturn(new com.baomidou.mybatisplus.core.metadata.IPage<AuditLog>(mockLogs, 0, 2, 2));
        
        target.queryAuditLogs(any(), any(), any(), any(), any(), any(), 1, 10);
    }

    @Test
    public void testDeleteAuditLogsBeforeDate() throws Exception {
        LocalDateTime cutoffDate = LocalDateTime.of(2023, 12, 31, 23, 59, 59);
        
        target.deleteAuditLogsBeforeDate(eq(cutoffDate));
        
        verify(auditService, times(1)).deleteAuditLogsBeforeDate(eq(cutoffDate));
    }

    @Test
    public void testExportAuditLogs() throws Exception {
        List<Long> ids = Arrays.asList(1L, 2L);
        
        target.exportAuditLogs(eq(ids), eq("csv"));
        
        verify(auditService, times(1)).exportAuditLogs(eq(ids), eq("csv"));
    }
}
