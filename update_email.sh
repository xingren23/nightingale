#!/bin/bash

####获取当前分支名称####
branch=`git rev-parse --abbrev-ref HEAD`

#####修改commit#######
git filter-branch -f --env-filter '
  
   OLD_EMAIL="原邮箱地址"
   CORRECT_NAME="新用户名"
   CORRECT_EMAIL="新邮箱地址"
   
   if [ "${GIT_COMMITTER_EMAIL}" = "${OLD_EMAIL}" ]
   then
      export GIT_COMMITTER_NAME="${CORRECT_NAME}"
      export GIT_COMMITTER_EMAIL="${CORRECT_EMAIL}"
   fi
  
   if [ "${GIT_AUTHOR_EMAIL}" = "${OLD_EMAIL}" ]
   then
       export GIT_AUTHOR_NAME="${CORRECT_NAME}"
       export GIT_AUTHOR_EMAIL="${CORRECT_EMAIL}"
   fi

' HEAD ^origin/${branch}
