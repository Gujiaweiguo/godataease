package io.dataease.menu.bo;

import io.dataease.menu.dao.auto.entity.CoreMenu;

import java.util.ArrayList;
import java.util.List;

public class MenuTreeNode extends CoreMenu {

    private List<MenuTreeNode> children = new ArrayList<>();

    public List<MenuTreeNode> getChildren() {
        return children;
    }

    public void setChildren(List<MenuTreeNode> children) {
        this.children = children;
    }
}
