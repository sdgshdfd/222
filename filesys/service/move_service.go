package service

import (
	"filesys/dao"
)

// 递归更新子文件路径
func UpdateChildPaths(tx *dao.Query, folderID, newParentID uint) error {
	// 获取所有子文件
	children, err := tx.File.
		Where(dao.File.ParentID.Eq(int32(folderID))).
		Find()
	if err != nil {
		return err
	}

	// 更新每个子文件的父目录
	for _, child := range children {
		// 更新子文件的父目录为新目录的ID
		if _, err := tx.File.
			Where(dao.File.ID.Eq(child.ID)).
			Update(dao.File.ParentID, int32(newParentID)); err != nil {
			return err
		}

		// 如果是文件夹，递归更新其子文件
		if child.Type == "folder" {
			if err := UpdateChildPaths(tx, uint(child.ID), newParentID); err != nil {
				return err
			}
		}
	}

	return nil
}
