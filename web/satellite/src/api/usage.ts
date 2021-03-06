// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

import apollo from '@/utils/apolloManager';
import gql from 'graphql-tag';

// fetchProjectUsage retrieves total project usage for a given period
export async function fetchProjectUsage(projectID: string, since: Date, before: Date): Promise<RequestResponse<ProjectUsage>> {
    let result: RequestResponse<ProjectUsage> = {
        errorMessage: '',
        isSuccess: false,
        data: {} as ProjectUsage
    };

    let response: any = await apollo.query(
        {
            query: gql(`
                query {
                    project(id: "${projectID}") {
                        usage(since: "${since.toISOString()}", before: "${before.toISOString()}") {
                            storage,
                            egress,
                            objectCount,
                            since,
                            before
                        }
                    }
                }`
            ),
            fetchPolicy: 'no-cache',
            errorPolicy: 'all'
        }
    );

    if (response.errors) {
        result.errorMessage = response.errors[0].message;
    } else {
        result.isSuccess = true;
        result.data = response.data.project.usage;
    }

    return result;
}

// fetchBucketUsages retrieves bucket usage totals for a particular project
export async function fetchBucketUsages(projectID: string, before: Date, cursor: BucketUsageCursor): Promise<RequestResponse<BucketUsagePage>> {
    let result: RequestResponse<BucketUsagePage> = {
        errorMessage: '',
        isSuccess: false,
        data: {} as BucketUsagePage
    };

    let response: any = null;
    try {
        response = await apollo.query(
            {
                query: gql(`
                    query {
                        project(id: "${projectID}") {
                            bucketUsages(before: "${before.toISOString()}", cursor: {
                                    limit: ${cursor.limit}, search: "${cursor.search}", page: ${cursor.page}
                                }) {
                                    bucketUsages{
                                        bucketName,
                                        storage,
                                        egress,
                                        objectCount,
                                        since,
                                        before
                                    },
                                    search,
                                    limit,
                                    offset,
                                    pageCount,
                                    currentPage,
                                    totalCount 
                            }
                        }
                    }`
                ),
                fetchPolicy: 'no-cache',
                errorPolicy: 'all'
            }
        );
    } catch (e) {
        console.log(e);
    }

    if (response.errors) {
        result.errorMessage = response.errors[0].message;
    } else {
        result.isSuccess = true;
        result.data = response.data.project.bucketUsages;
    }

    return result;
}
