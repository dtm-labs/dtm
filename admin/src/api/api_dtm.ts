import { AxiosResponse } from 'axios'
import request from '/@/utils/request'

export interface IListAllTransactionsReq {
    gid?: string;
    limit: number;
    position?: string;
}

export interface IListAllKVReq {
    cat: string;
    limit: number;
    position?: string;
}

export function listAllTransactions<T>(
    payload: IListAllTransactionsReq
): Promise<AxiosResponse<T>> {
    return request({
        url: '/api/dtmsvr/all',
        method: 'get',
        params: payload
    })
}

export function forceStopTransaction(gid: string): Promise<AxiosResponse> {
    return request({
        url: '/api/dtmsvr/forceStop',
        method: 'post',
        data: { gid }
    })
}

export function queryKVPair<T>(payload: {
    cat: string;
    key: string;
}): Promise<AxiosResponse<T>> {
    return request({
        url: '/api/dtmsvr/queryKV',
        method: 'get',
        params: payload
    })
}

export function listKVPairs<T>(
    payload: IListAllKVReq
): Promise<AxiosResponse<T>> {
    return request({
        url: '/api/dtmsvr/scanKV',
        method: 'get',
        params: payload
    })
}

export function deleteTopic<T>(topicName: string): Promise<AxiosResponse<T>> {
    return request({
        url: `/api/dtmsvr/topic/${topicName}`,
        method: 'delete'
    })
}

export function subscribe<T>(payload: {
    topic: string;
    url: string;
    remark: string;
}): Promise<AxiosResponse<T>> {
    return request({
        url: '/api/dtmsvr/subscribe',
        method: 'get',
        params: payload
    })
}

export function unsubscribe(payload: {
    topic: string;
    url: string;
}): Promise<AxiosResponse> {
    return request({
        url: '/api/dtmsvr/unsubscribe',
        method: 'get',
        params: payload
    })
}

export function getTransaction<T>(payload: {
    gid: string;
}): Promise<AxiosResponse<T>> {
    return request({
        url: '/api/dtmsvr/query',
        method: 'get',
        params: payload
    })
}

export function resetNextCronTime(gid: string): Promise<AxiosResponse> {
    return request({
        url: '/api/dtmsvr/resetNextCronTime',
        method: 'post',
        data: { gid }
    })
}

export function getDtmVersion(): Promise<AxiosResponse<any>> {
    return request({
        url: '/api/dtmsvr/version',
        method: 'get'
    })
}
