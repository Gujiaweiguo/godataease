package io.dataease.audit.dao.auto.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import io.dataease.audit.entity.LoginFailure;
import org.apache.ibatis.annotations.Mapper;

@Mapper
public interface LoginFailureMapper extends BaseMapper<LoginFailure> {
}
