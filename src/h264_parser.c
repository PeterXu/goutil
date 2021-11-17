#include <stdint.h>
#include <string.h>
#include <stdbool.h>

static uint32_t h264_eg_getbit(uint8_t *base, uint32_t offset) {
    return ((*(base + (offset >> 0x3))) >> (0x7 - (offset & 0x7))) & 0x1;
}

static uint32_t h264_eg_getbits(uint8_t *base, uint32_t *offset, int num) {
    uint32_t res = 0; 
    for(int32_t i=num-1; i>=0; i--) {
        res |= h264_eg_getbit(base, (*offset)++) << i;
    }    
    return res; 
}

static uint32_t h264_eg_decode(uint8_t *base, uint32_t *offset) {
    uint32_t zeros = 0; 
    while(h264_eg_getbit(base, (*offset)++) == 0)
        zeros++;
    uint32_t res = 1 << zeros;
    int32_t i = 0; 
    for(i=zeros-1; i>=0; i--) {
        res |= h264_eg_getbit(base, (*offset)++) << i;
    }    
    return res-1;
}

static void h264_parse_rbsp(const uint8_t *data, size_t length, uint8_t *rbsp_buffer, size_t *outlen) {
    size_t rbsp_len = 0;
    for (size_t i = 0; i < length;) {
        if (length - i >= 3 && data[i] == 0 && data[i + 1] == 0 && data[i + 2] == 3) {
            memcpy(rbsp_buffer+rbsp_len, data+i, 2);
            rbsp_len += 2;
            i += 3;
        }else {
            memcpy(rbsp_buffer+rbsp_len, data+i, 1);
            rbsp_len += 1;
            ++i;
        }
    }
    *outlen = rbsp_len;
}

static bool h264_parse_pps(const uint8_t *data, size_t length, uint32_t *p_pps_id, uint32_t *p_sps_id) {
    uint8_t *rbsp_buffer = (uint8_t *)data+1;
    size_t rbsp_len = length;

    uint32_t offset = 0;
    uint8_t *base = (uint8_t *)(rbsp_buffer);
    uint32_t pps_id = h264_eg_decode(base, &offset);
    uint32_t sps_id = h264_eg_decode(base, &offset);
    if (p_pps_id) *p_pps_id = pps_id;
    if (p_sps_id) *p_sps_id = sps_id;
    // printf("pps nalu, pps_id=%u, sps_id=%u", pps_id, sps_id);
    return true;
}

static bool h264_parse_slice_pps(const uint8_t *data, size_t length, uint32_t *p_pps_id) {
    uint8_t *rbsp_buffer = (uint8_t *)data+1;
    size_t rbsp_len = length;

    uint32_t golomb_tmp = 0;
    uint32_t offset = 0;
    uint8_t *base = (uint8_t *)(rbsp_buffer);
    // first_mb_in_slice: ue(v)
    golomb_tmp = h264_eg_decode(base, &offset);
    // slice_type: ue(v)
    golomb_tmp = h264_eg_decode(base, &offset);
    // pic_parameter_set_id: ue(v)
    uint32_t slice_pps_id = h264_eg_decode(base, &offset);
    if (p_pps_id) *p_pps_id = slice_pps_id;
    // printf("pps slice_pps_id=%u", slice_pps_id);
    return true;
}

static bool h264_parse_sps(const uint8_t *buffer, size_t length, int *p_width, int *p_height, uint32_t *p_sps_id) {
    uint8_t *rbsp_buffer = (uint8_t *)buffer;
    size_t rbsp_len = length;

    uint32_t chroma_format_idc = 1;
    uint32_t separate_colour_plane_flag = 0;

    /* Let's check if it's the right profile, first */
    int index = 1;
    int profile_idc = *(rbsp_buffer+index);
    index += 1;
    /* Then let's skip 2 bytes and evaluate/skip the rest */
    index += 2;
    uint32_t offset = 0;
    uint8_t *base = (uint8_t *)(rbsp_buffer+index);
    /* Skip seq_parameter_set_id */
    uint32_t sps_id = h264_eg_decode(base, &offset);
    if (p_sps_id) *p_sps_id = sps_id;
    //printf("sps nal, sps_id=%u, profle=%d", sps_id, profile_idc);

    if(profile_idc != 66) {
        //printf("Profile is not baseline, profile_idc=%d", profile_idc);
        if (profile_idc == 100 || profile_idc == 110 || profile_idc == 122 ||
                profile_idc == 244 || profile_idc == 44 || profile_idc == 83 ||
                profile_idc == 86 || profile_idc == 118 || profile_idc == 128 ||
                profile_idc == 138 || profile_idc == 139 || profile_idc == 134) {
            // chroma_format_idc: ue(v)
            chroma_format_idc = h264_eg_decode(base, &offset);
            if (chroma_format_idc == 3) {
                // separate_colour_plane_flag: u(1)
                separate_colour_plane_flag = h264_eg_getbit(base, offset++);
            }
            // bit_depth_luma_minus8: ue(v)
            h264_eg_decode(base, &offset);
            // bit_depth_chroma_minus8: ue(v)
            h264_eg_decode(base, &offset);
            // qpprime_y_zero_transform_bypass_flag: u(1)
            h264_eg_getbit(base, offset++);
            // seq_scaling_matrix_present_flag: u(1)
            uint32_t seq_scaling_matrix_present_flag = h264_eg_getbit(base, offset++);
            if (seq_scaling_matrix_present_flag) {
                uint32_t seq_scaling_list_present_flags;
                if (chroma_format_idc != 3) {
                    h264_eg_getbits(base, &offset, 8);
                }else {
                    h264_eg_getbits(base, &offset, 12);
                }
                if (seq_scaling_list_present_flags > 0) {
                    // printf("SPS contains scaling lists, which are unsupported.");
                    return false;
                }
            }
        }else {
            //printf("unsupported profileidc=%d", profile_idc);
            return false;
        }
    }
    /* Skip log2_max_frame_num_minus4 */
    h264_eg_decode(base, &offset);
    /* Evaluate pic_order_cnt_type */
    int pic_order_cnt_type = h264_eg_decode(base, &offset);
    if(pic_order_cnt_type == 0) {
        /* Skip log2_max_pic_order_cnt_lsb_minus4 */
        h264_eg_decode(base, &offset);
    } else if(pic_order_cnt_type == 1) {
        /* Skip delta_pic_order_always_zero_flag, offset_for_non_ref_pic,
         * offset_for_top_to_bottom_field and num_ref_frames_in_pic_order_cnt_cycle */
        h264_eg_getbit(base, offset++);
        h264_eg_decode(base, &offset);
        h264_eg_decode(base, &offset);
        int num_ref_frames_in_pic_order_cnt_cycle = h264_eg_decode(base, &offset);
        int i = 0;
        for(i=0; i<num_ref_frames_in_pic_order_cnt_cycle; i++) {
            h264_eg_decode(base, &offset);
        }
    }
    /* Skip max_num_ref_frames and gaps_in_frame_num_value_allowed_flag */
    h264_eg_decode(base, &offset);
    h264_eg_getbit(base, offset++);
    /* We need the following three values */
    int pic_width_in_mbs_minus1 = h264_eg_decode(base, &offset);
    int pic_height_in_map_units_minus1 = h264_eg_decode(base, &offset);
    int frame_mbs_only_flag = h264_eg_getbit(base, offset++);
    if(!frame_mbs_only_flag) {
        /* Skip mb_adaptive_frame_field_flag */
        h264_eg_getbit(base, offset++);
    }
    /* Skip direct_8x8_inference_flag */
    h264_eg_getbit(base, offset++);
    /* We need the following value to evaluate offsets, if any */
    int frame_cropping_flag = h264_eg_getbit(base, offset++);
    int frame_crop_left_offset = 0, frame_crop_right_offset = 0,
        frame_crop_top_offset = 0, frame_crop_bottom_offset = 0;
    if(frame_cropping_flag) {
        frame_crop_left_offset = h264_eg_decode(base, &offset);
        frame_crop_right_offset = h264_eg_decode(base, &offset);
        frame_crop_top_offset = h264_eg_decode(base, &offset);
        frame_crop_bottom_offset = h264_eg_decode(base, &offset);
    }
    /* Skip vui_parameters_present_flag */
    h264_eg_getbit(base, offset++);

    if (separate_colour_plane_flag || chroma_format_idc == 0) {
        frame_crop_bottom_offset *= (2 - frame_mbs_only_flag);
        frame_crop_top_offset *= (2 - frame_mbs_only_flag);
    }else if (!separate_colour_plane_flag && chroma_format_idc > 0){
        if (chroma_format_idc == 1 || chroma_format_idc == 2) {
            frame_crop_left_offset *= 2;
            frame_crop_right_offset *= 2;
        }
        if (chroma_format_idc == 1) {
            frame_crop_top_offset *= 2;
            frame_crop_bottom_offset *= 2;
        }
    }

    /* We skipped what we didn't care about and got what we wanted, compute width/height */
    if(p_width)
        *p_width = ((pic_width_in_mbs_minus1 +1)*16) - (frame_crop_left_offset + frame_crop_right_offset);
    if(p_height)
        *p_height = ((2 - frame_mbs_only_flag)* (pic_height_in_map_units_minus1 +1) * 16) - (frame_crop_top_offset + frame_crop_bottom_offset);
    return true;
}

bool parse_rtp_video(const uint8_t* nal_data, size_t nal_size, uint8_t *p_rtp_type,
        int *p_width, int *p_height, uint32_t *p_sps_id) {
    if (nal_size < 4 || nal_size >= 1500) {
        return false;
    }

    int width = 0;
    int height = 0;
    uint32_t sps_id = 32;

    bool bfind = false;
    uint8_t nal_type = nal_data[0] & 0x1f;
    if (p_rtp_type) {
        *p_rtp_type = nal_type;
    }

    if (nal_type == 7) { // sps
        bfind = h264_parse_sps(nal_data, nal_size, &width, &height, &sps_id);
    }else if (nal_type == 24) {
        int tot = nal_size - 1 ;
        const uint8_t *buffer = nal_data + 1;
        while(tot > 0) {
            uint16_t psize = 0;
            psize = (buffer[0]<<8) + buffer[1];
            buffer += 2;
            tot -= 2;

            nal_type = buffer[0] & 0x1f;
            if (nal_type == 7) {
                bfind = h264_parse_sps(buffer, psize, &width, &height, &sps_id);
            }
            buffer += psize;
            tot -= psize;
        }
    }

    if (bfind && p_width && p_height) {
        *p_width = width;
        *p_height = height;
        if (p_sps_id) {
            *p_sps_id = sps_id;
        }
    }
    return bfind;
}
