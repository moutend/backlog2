// Code generated by "stringer -type=cacheType"; DO NOT EDIT.

package main

import "strconv"

const _cacheType_name = "issueCommentsCachePathissueTypesCachePathissuesCachePathissueCachePathmyselfCachePathprioritiesCachePathprojectsCachePathprojectCachePathpullRequestsCachePathrepositoriesCachePathstatusesCachePathwikisCachePathwikiCachePath"

var _cacheType_index = [...]uint8{0, 22, 41, 56, 70, 85, 104, 121, 137, 158, 179, 196, 210, 223}

func (i cacheType) String() string {
	if i < 0 || i >= cacheType(len(_cacheType_index)-1) {
		return "cacheType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _cacheType_name[_cacheType_index[i]:_cacheType_index[i+1]]
}
