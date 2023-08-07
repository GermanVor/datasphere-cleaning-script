package cleaningScript

import (
	"fmt"
	"sync"
	"time"

	"github.com/GermanVor/datasphere-cleaning-script/common"
	"github.com/GermanVor/datasphere-cleaning-script/service/datasphere"
)

type LogbookValue struct {
	projectsCount            int32
	deletedProjectsCount     int32
	errorDeletedProjectCount int32
	community                datasphere.Community
}

type CommunityLogbook struct {
	m  map[string]*LogbookValue
	mu sync.RWMutex
}

func Run(datasphereClient *datasphere.Client, ORGANIZATION_ID, COMMUNITY_SUBSTR string) {
	common.Log("Cleaning Script Started")

	communityArr := datasphereClient.GetCommunities(ORGANIZATION_ID, COMMUNITY_SUBSTR)
	communityArrLen := len(communityArr)

	COMMUNITY_LOGBOOK := CommunityLogbook{
		m:  make(map[string]*LogbookValue),
		mu: sync.RWMutex{},
	}

	communitiesCh := make(chan datasphere.Community)
	projectsCh := make(chan datasphere.Project)

	common.Log(fmt.Sprintf("Communities number is %d", communityArrLen))

	communitiesWG := sync.WaitGroup{}
	communitiesWG.Add(communityArrLen)

	go func() {
		ret := common.Retry{
			RetryLimit: 3,
		}

		deleteCommunity := common.Debounce(func(community datasphere.Community) {
			defer communitiesWG.Done()
			common.Log(fmt.Sprintf("\tDeleting Community %s", community.Id))

			err := ret.Call(func() error {
				return datasphereClient.DeleteCommunity(community.Id)
			})

			if err == nil {
				common.Log(fmt.Sprintf("\tCommunity %s was deleted", community.Id))
			} else {
				common.Log(fmt.Sprintf("\tCommunity %s was not deleted cause %s", community.Id, err))
			}
		}, time.Second)

		for community := range communitiesCh {
			go deleteCommunity(community)
		}
	}()

	go func() {
		ret := common.Retry{
			RetryLimit: 5,
		}

		deleteProject := common.Debounce(func(project datasphere.Project) {
			if mValue, ok := COMMUNITY_LOGBOOK.m[project.CommunityId]; ok {
				common.Log(fmt.Sprintf("\tDeleting Project %s from Community %s", project.Id, project.CommunityId))

				err := ret.Call(func() error {
					return datasphereClient.DeleteProject(project.Id)
				})

				COMMUNITY_LOGBOOK.mu.Lock()
				defer COMMUNITY_LOGBOOK.mu.Unlock()

				mValue.projectsCount--
				if err == nil {
					common.Log(fmt.Sprintf("\tProject %s was deleted", project.Id))
					mValue.deletedProjectsCount++
				} else {
					common.Log(fmt.Sprintf("\tProject %s was not deleted cause %s", project.Id, err))
					mValue.errorDeletedProjectCount++
				}

				if mValue.projectsCount == 0 {
					if mValue.errorDeletedProjectCount == 0 {
						communitiesCh <- mValue.community
					} else {
						communitiesWG.Done()
					}
				}
			}
		}, time.Second+500*time.Millisecond)

		for project := range projectsCh {
			go deleteProject(project)
		}
	}()

	for _, community := range communityArr {
		projectArr, _ := datasphereClient.GetProjects(community.Id)
		projectArrLen := int32(len(projectArr))

		COMMUNITY_LOGBOOK.m[community.Id] = &LogbookValue{
			projectsCount:            projectArrLen,
			deletedProjectsCount:     0,
			errorDeletedProjectCount: 0,
			community:                community,
		}

		if projectArrLen == 0 {
			communitiesCh <- community
			continue
		}

		for _, project := range projectArr {
			projectsCh <- project
		}
	}

	communitiesWG.Wait()
	close(communitiesCh)
	close(projectsCh)

	common.Log("Cleaning Script Finished")
	common.Log("Statistic - (total|success deleted|error deleted) Projects")

	deletedCommunityCount := 0
	for _, v := range COMMUNITY_LOGBOOK.m {
		common.Log(fmt.Sprintf(
			"\tCommunity %s (%d|%d|%d)",
			v.community.Id,
			v.deletedProjectsCount+v.errorDeletedProjectCount,
			v.deletedProjectsCount,
			v.errorDeletedProjectCount,
		))

		if v.projectsCount == 0 && v.errorDeletedProjectCount == 0 {
			deletedCommunityCount++
		}
	}

	common.Log(fmt.Sprintf("Community deleted %d/%d", communityArrLen, deletedCommunityCount))
}
