package io.dataease.audit;

import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import io.dataease.audit.entity.AuditLog;
import io.dataease.audit.server.AuditController;
import io.dataease.audit.service.IAuditService;
import io.dataease.result.ResultMessage;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyInt;
import static org.mockito.ArgumentMatchers.anyLong;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
public class AuditControllerTest {

    @Mock
    private IAuditService auditService;

    @Test
    public void listAuditLogsReturnsOk() {
        Page<AuditLog> page = new Page<>(1, 10);
        when(auditService.queryAuditLogs(
            any(), any(), any(), any(), any(),
            any(), any(), any(), any()
        )).thenReturn(page);

        AuditController controller = new AuditController(auditService);
        ResultMessage result = controller.queryAuditLogs(
            null, null, null, null, null, null, null, 1, 10
        );

        assertNotNull(result);
        assertEquals(0, result.getCode());
    }
}
